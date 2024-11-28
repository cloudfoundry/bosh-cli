package cmd

import (
	"fmt"
	"unicode"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	"github.com/cloudfoundry/bosh-cli/v7/pcap"
)

const (
	maxDeviceNameLength = 16
	maxFilterLength     = 5000
)

type PcapCmd struct {
	deployment boshdir.Deployment
	pcapRunner pcap.PcapRunner
	parallel   int
}

func NewPcapCmd(
	deployment boshdir.Deployment,
	pcapRunner pcap.PcapRunner,
	parallel int,
) PcapCmd {
	return PcapCmd{
		deployment: deployment,
		pcapRunner: pcapRunner,
		parallel:   parallel,
	}
}

func (c PcapCmd) Run(opts PcapOpts) error {
	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	var result boshdir.SSHResult

	// If no slugs are provided, default to capturing all instances by using an empty slug.
	slugs := []boshdir.AllOrInstanceGroupOrInstanceSlug{{}}

	if len(opts.Args.Slugs) > 0 {
		slugs = opts.Args.Slugs
	}

	slugs = boshdir.DeduplicateSlugs(slugs)

	for _, slug := range slugs {
		res, err := c.deployment.SetUpSSH(slug, sshOpts)
		if err != nil {
			return err
		}

		result.Hosts = append(result.Hosts, res.Hosts...)
	}

	defer func() {
		for _, slug := range slugs {
			_ = c.deployment.CleanUpSSH(slug, sshOpts)
		}
	}()

	argv, err := buildPcapCmd(opts)
	if err != nil {
		return fmt.Errorf("invalid pcap cmd options: %w", err)
	}

	return c.pcapRunner.Run(result, sshOpts.Username, argv, opts, connOpts.PrivateKey, c.parallel)
}

func buildPcapCmd(opts PcapOpts) (string, error) {
	err := validateDevice(opts.Interface)
	if err != nil {
		return "", err
	}

	if len(opts.Filter) > maxFilterLength {
		return "", fmt.Errorf("expected filter to be at most %d characters, received %d", maxFilterLength, len(opts.Filter))
	}

	return fmt.Sprintf("sudo tcpdump -w - -i %s -s %d", opts.Interface, opts.SnapLength), nil
}

// validateDevice is a go implementation of dev_valid_name from the linux kernel.
//
// See: https://lxr.linux.no/linux+v6.0.9/net/core/dev.c#L995
func validateDevice(name string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("validate network interface name: %w", err)
		}
	}()

	if len(name) > maxDeviceNameLength {
		return fmt.Errorf("name too long: %d > %d", len(name), maxDeviceNameLength)
	}

	if name == "." || name == ".." {
		return fmt.Errorf("invalid name: '%s'", name)
	}

	for i, r := range name {
		if r == '/' {
			return fmt.Errorf("%w at pos. %d: '/'", pcap.ErrIllegalCharacter, i)
		}
		if r == '\x00' {
			return fmt.Errorf("%w at pos. %d: '\\0'", pcap.ErrIllegalCharacter, i)
		}
		if r == ':' {
			return fmt.Errorf("%w at pos. %d: ':'", pcap.ErrIllegalCharacter, i)
		}
		if unicode.Is(unicode.White_Space, r) {
			return fmt.Errorf("%w: whitespace at pos %d", pcap.ErrValidationFailed, i)
		}
	}

	return nil
}
