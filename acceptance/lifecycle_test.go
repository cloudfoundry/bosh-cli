package acceptance_test

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

var _ = Describe("bosh-micro", func() {
	var (
		cmdRunner acceptanceCmdRunner
		testEnv   acceptanceEnvironment
	)

	BeforeSuite(func() {
		localEnv, err := parseEnv()
		Expect(err).NotTo(HaveOccurred())

		logger := boshlog.NewLogger(boshlog.LevelDebug)
		fileSystem := boshsys.NewOsFileSystem(logger)
		testEnv = newRemoteTestEnvironment(
			localEnv.vmUsername,
			localEnv.vmIP,
			localEnv.privateKeyPath,
			fileSystem,
			logger,
		)

		cmdRunner = NewSSHCmdRunner(
			localEnv.vmUsername,
			localEnv.vmIP,
			localEnv.privateKeyPath,
			logger,
		)

		err = bmtestutils.BuildExecutableForArch("linux-amd64")
		Expect(err).NotTo(HaveOccurred())

		boshMicroPath := "./../out/bosh-micro"
		Expect(fileSystem.FileExists(boshMicroPath)).To(BeTrue())
		err = testEnv.Copy("bosh-micro", boshMicroPath)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("stemcell", localEnv.stemcellURL)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("cpiRelease", localEnv.cpiReleaseURL)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is able to deploy a CPI release with a stemcell", func() {
		contents := `---
name: test-release
resource_pools:
- name: fake-resource-pool-name
  env:
    bosh:
      password: secret
networks:
- name: fake-network-name
  type: dynamic
  cloud_properties:
    subnet: fake-subnet
    a:
      b: value
cloud_provider:
  properties:
    cpi:
      warden:
        connect_network: tcp
        connect_address: 0.0.0.0:7777
        network_pool: 10.244.0.0/16
        host_ip: 192.168.54.4
      agent:
        mbus: 192.168.54.4
`
		testEnv.WriteContentString("manifest", contents)

		_, _, exitCode, err := cmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deployment", testEnv.Path("manifest"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		stdout, _, exitCode, err := cmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		Expect(stdout).To(ContainSubstring("uploading stemcell"))
		Expect(stdout).To(ContainSubstring("Creating VM from"))
	})
})

type acceptanceEnvironment interface {
	Path(string) string
	Copy(string, string) error
	WriteContentString(string, string) error
	RemoteDownload(string, string) error
	DownloadOrCopy(string, string) error
}

type remoteTestEnvironment struct {
	vmUsername     string
	vmIP           string
	privateKeyPath string
	cmdRunner      boshsys.CmdRunner
	fileSystem     boshsys.FileSystem
}

func newRemoteTestEnvironment(
	vmUsername string,
	vmIP string,
	privateKeyPath string,
	fileSystem boshsys.FileSystem,
	logger boshlog.Logger,
) remoteTestEnvironment {
	return remoteTestEnvironment{
		vmUsername:     vmUsername,
		vmIP:           vmIP,
		privateKeyPath: privateKeyPath,
		cmdRunner:      boshsys.NewExecCmdRunner(logger),
		fileSystem:     fileSystem,
	}
}

func (e remoteTestEnvironment) Path(name string) string {
	return path.Join("/", "home", e.vmUsername, name)
}

func (e remoteTestEnvironment) Copy(destName, srcPath string) error {
	if srcPath == "" {
		return fmt.Errorf("Cannot use an empty file for `%s'", destName)
	}

	_, _, exitCode, err := e.cmdRunner.RunCommand(
		"scp",
		"-o", "StrictHostKeyChecking=no",
		"-i", e.privateKeyPath,
		srcPath,
		fmt.Sprintf("%s@%s:%s", e.vmUsername, e.vmIP, e.Path(destName)),
	)
	if exitCode != 0 {
		return fmt.Errorf("scp of `%s' to `%s' failed", srcPath, destName)
	}
	return err
}

func (e remoteTestEnvironment) DownloadOrCopy(destName, src string) error {
	if strings.HasPrefix(src, "http") {
		return e.RemoteDownload(destName, src)
	}
	return e.Copy(destName, src)
}

func (e remoteTestEnvironment) RemoteDownload(destName, srcURL string) error {
	if srcURL == "" {
		return fmt.Errorf("Cannot use an empty file for `%s'", destName)
	}

	_, _, exitCode, err := e.cmdRunner.RunCommand(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", e.privateKeyPath,
		fmt.Sprintf("%s@%s", e.vmUsername, e.vmIP),
		fmt.Sprintf("wget -q -O %s %s", destName, srcURL),
	)
	if exitCode != 0 {
		return fmt.Errorf("download of `%s' to `%s' failed", srcURL, destName)
	}
	return err
}

func (e remoteTestEnvironment) WriteContentString(destName, contents string) error {
	tmpFile, err := e.fileSystem.TempFile("bosh-micro-cli-acceptance")
	if err != nil {
		return err
	}
	defer e.fileSystem.RemoveAll(tmpFile.Name())
	_, err = tmpFile.WriteString(contents)
	if err != nil {
		return err
	}
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	return e.Copy(destName, tmpFile.Name())
}

type acceptanceCmdRunner interface {
	RunCommand(...string) (string, string, int, error)
}

type sshCmdRunner struct {
	vmUsername     string
	vmIP           string
	privateKeyPath string
	runner         boshsys.CmdRunner
}

func NewSSHCmdRunner(
	vmUsername string,
	vmIP string,
	privateKeyPath string,
	logger boshlog.Logger,
) sshCmdRunner {
	return sshCmdRunner{
		vmUsername:     vmUsername,
		vmIP:           vmIP,
		privateKeyPath: privateKeyPath,
		runner:         boshsys.NewExecCmdRunner(logger),
	}
}

func (r sshCmdRunner) RunCommand(args ...string) (string, string, int, error) {
	argsWithEnv := append([]string{fmt.Sprintf("TMPDIR=/home/%s", r.vmUsername)}, args...)
	return r.runner.RunCommand(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", r.privateKeyPath,
		fmt.Sprintf("%s@%s", r.vmUsername, r.vmIP),
		strings.Join(argsWithEnv, " "),
	)
}

type localEnvironment struct {
	vmUsername     string
	vmIP           string
	privateKeyPath string
	stemcellURL    string
	cpiReleaseURL  string
}

func parseEnv() (localEnvironment, error) {
	env := localEnvironment{
		vmUsername:     os.Getenv("BOSH_MICRO_VM_USERNAME"),
		vmIP:           os.Getenv("BOSH_MICRO_VM_IP"),
		privateKeyPath: os.Getenv("BOSH_MICRO_PRIVATE_KEY"),
		stemcellURL:    os.Getenv("BOSH_MICRO_STEMCELL"),
		cpiReleaseURL:  os.Getenv("BOSH_MICRO_CPI_RELEASE"),
	}

	var err error
	if env.vmUsername == "" {
		fmt.Println("BOSH_MICRO_VM_USERNAME must be set")
		err = errors.New("")
	}
	if env.vmIP == "" {
		fmt.Println("BOSH_MICRO_VM_IP must be set")
		err = errors.New("")
	}
	if env.privateKeyPath == "" {
		fmt.Println("BOSH_MICRO_PRIVATE_KEY must be set")
		err = errors.New("")
	}
	if env.stemcellURL == "" {
		fmt.Println("BOSH_MICRO_STEMCELL must be set")
		err = errors.New("")
	}
	if env.cpiReleaseURL == "" {
		fmt.Println("BOSH_MICRO_CPI_RELEASE must be set")
		err = errors.New("")
	}

	return env, err
}
