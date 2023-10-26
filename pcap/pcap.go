package pcap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
	"github.com/gopacket/gopacket/pcapgo"
	"golang.org/x/crypto/ssh"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

var (
	ErrValidationFailed = fmt.Errorf("validation failed")
	ErrIllegalCharacter = fmt.Errorf("illegal character: %w", ErrValidationFailed)
	openHandleError     = fmt.Errorf("unable to open pcap handle")
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . PcapRunner
type PcapRunner interface {
	Run(boshdir.SSHResult, string, string, PcapOpts, ssh.Signer) error
}

func NewPcapRunner(ui boshui.UI) PcapRunner {
	return PcapRunnerImpl{ui: ui}
}

type PcapRunnerImpl struct {
	ui boshui.UI
}

func (p PcapRunnerImpl) Run(result boshdir.SSHResult, username string, argv string, opts PcapOpts, privateKey ssh.Signer) error {
	var packetCs []<-chan gopacket.Packet

	done := make(chan struct{})

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancelCause(context.Background())

	runningCaptures := 0

	for _, host := range result.Hosts {

		p.ui.BeginLinef("Start capture on %s/%s\n", host.Job, host.IndexOrID)

		packets, err := captureSSH(argv, opts.Filter, username, host, privateKey, opts.StopTimeout, wg, done, p.ui, ctx, cancel)
		if err != nil {
			// c.ui.ErrorLinef writes error message to stdout/sdterr but does not stop the workflow
			p.ui.ErrorLinef("Capture cannot be started on the instance %s/%s due to error: %s. \nContinue on other instances", host.Job, host.IndexOrID, err.Error())

			continue
		}

		runningCaptures++

		packetCs = append(packetCs, packets)
	}

	if runningCaptures == 0 {
		return fmt.Errorf("starting of all pcap captures failed")
	}

	err := writePacketsToFile(opts.SnapLength, opts.Output, packetCs, p.ui)
	if err != nil {
		return fmt.Errorf("write to output file failed: %w", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case <-signals:
		close(done)
	case <-ctx.Done():
		// ctx canceled as cmd exited or an error occurred
	}

	p.ui.BeginLinef("Wait for session has been finished\n")

	wg.Wait()

	p.ui.EndLinef("\nCapture finished")

	return nil
}

func writePacketsToFile(snapLength uint32, outputFile string, packetCs []<-chan gopacket.Packet, ui boshui.UI) error {
	// setup output pcap-file
	packetFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	packetWriter := pcapgo.NewWriter(packetFile)
	err = packetWriter.WriteFileHeader(snapLength, layers.LinkTypeEthernet)
	if err != nil {
		return err
	}

	mergedPackets := mergePackets(packetCs)
	go func() {
		for packet := range mergedPackets {
			err = packetWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
			if err != nil {
				ui.ErrorLinef("Writing packet to file failed due to error: %s/n", err.Error())
			}
		}
		_ = packetFile.Sync()
		_ = packetFile.Close()
	}()
	return nil
}

func addFilterToCmd(tcpdump, filter, clientIP string, clientSSHPort int) string {
	filter = strings.TrimSpace(filter)
	if filter != "" {
		filter = fmt.Sprintf("not (host %s and port %d) and (%s)", clientIP, clientSSHPort, filter)
	} else {
		filter = fmt.Sprintf("not (host %s and port %d)", clientIP, clientSSHPort)
	}
	return fmt.Sprintf("%s %q", tcpdump, filter)
}

func captureSSH(tcpdumpCmd, filter, user string, host boshdir.Host, privateKeyPem ssh.Signer, stopTimeout time.Duration, wg *sync.WaitGroup, done chan struct{}, ui boshui.UI, ctx context.Context, cancel context.CancelCauseFunc) (<-chan gopacket.Packet, error) {
	client, err := ssh.Dial("tcp", host.Host+":22", &ssh.ClientConfig{
		Config:          ssh.Config{},
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(privateKeyPem)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("ssh: dial: %w", err)
	}

	clientSSHAddr, err := getSSHClientIP(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("outbound IP not found %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("ssh: new session: %w", err)
	}

	tcpdump := addFilterToCmd(tcpdumpCmd, filter, clientSSHAddr.IP.String(), clientSSHAddr.Port)
	ui.ErrorLinef(tcpdump)

	packets, err := openPcapHandle(tcpdump, session, wg, cancel)
	if err != nil {
		session.Close()
		client.Close()
		return nil, err
	}

	wg.Add(1)
	go func() {
		defer client.Close()
		defer wg.Done()

		ui.BeginLinef("\nRunning on %s/%s. To stop capturing traffic and generate a pcap file, press CTRL-C during the capture\n", host.Job, host.IndexOrID)

		select {
		case <-ctx.Done():
			// command exited, no need to send SIGTERM to terminate it
			err = ctx.Err()
			if err != nil && errors.Is(err, openHandleError) {
				ui.ErrorLinef("Capture stopped with error on %s/%s. No data written.", host.Job, host.IndexOrID)
				return
			}
		case <-done:
			// if termination signal will be sent by user
			ui.EndLinef("Stop capture on %s/%s, waiting %.0f seconds for data to flush", host.Job, host.IndexOrID, stopTimeout.Seconds())

			err := session.Signal(ssh.SIGTERM)
			if err != nil {
				ui.ErrorLinef("Unable to tell tcpdump to stop: %s\n", err.Error())
			}
		}

		time.Sleep(stopTimeout)

		_ = session.Close()
	}()

	return packets, nil
}

func getSSHClientIP(client *ssh.Client) (*net.TCPAddr, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh: new session: %w", err)
	}

	defer session.Close()

	output, err := session.Output("bash -c 'declare -a SSH; SSH=( $SSH_CONNECTION ); echo ${SSH[0]}:${SSH[1]}'")
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", strings.TrimSpace(string(output)))
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func openPcapHandle(tcpdumpCmd string, session *ssh.Session, wg *sync.WaitGroup, cancel context.CancelCauseFunc) (<-chan gopacket.Packet, error) {
	readable, writeable, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("os: pipe: %w", err)
	}
	session.Stdout = writeable

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		_, _ = io.Copy(os.Stderr, stderr)
	}()

	// The session must start before we open the handle, otherwise opening the
	// handle blocks forever. Probably since it expects to be able to read the
	// header information as we are opening an offline file from its POV.
	err = session.Start(tcpdumpCmd)
	if err != nil {
		return nil, fmt.Errorf("ssh: start session: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// waits for remote command to exit
		err = session.Wait()
		if err != nil {
			fmt.Println("ssh session died:", err.Error())
			cancel(err)

			writeable.Close()
			readable.Close()
		}
	}()

	// Start the separate goroutine for a pcap handle creation and receiving of the packets.
	// The pcap handle can be created without blocking after readable exists.
	out := make(chan gopacket.Packet)
	go func() {
		handle, err := pcap.OpenOfflineFile(readable)
		if err != nil {
			cancel(openHandleError)
			return
		}

		packetSource := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)

		// Run a goroutine that will write packets to the channel
		go func() {
			defer handle.Close()

			for packet := range packetSource.Packets() {
				out <- packet
			}
		}()
	}()
	return out, nil
}

func mergePackets(packetCs []<-chan gopacket.Packet) <-chan gopacket.Packet {
	// Taken from: https://go.dev/blog/pipelines#fan-out-fan-in
	wg := &sync.WaitGroup{}
	out := make(chan gopacket.Packet)

	wg.Add(len(packetCs))
	for _, c := range packetCs {
		go func(c <-chan gopacket.Packet) {
			defer wg.Done()
			for p := range c {
				out <- p
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
