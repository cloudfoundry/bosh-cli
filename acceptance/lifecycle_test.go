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
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		fileSystem := boshsys.NewOsFileSystem(logger)
		testEnv = newVagrantTestEnvironment(fileSystem)

		cmdRunner = NewVagrantSSHCmdRunner(logger)
		err := cmdRunner.StartSession()
		Expect(err).ToNot(HaveOccurred())

		err = bmtestutils.BuildExecutableForArch("linux-amd64")
		Expect(err).NotTo(HaveOccurred())

		err = testEnv.Copy("bosh-micro", "./../out/bosh-micro")
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.Copy("stemcell", os.Getenv("BOSH_STEMCELL"))
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.Copy("cpiRelease", os.Getenv("BOSH_CPI_RELEASE"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		testEnv.Cleanup()
		if suiteFailed {
			println("\n\nUse `vagrant ssh' to get into vagrant VM to debug the failed tests")
		} else {
			cmdRunner.EndSession()
		}
	})

	It("is able to deploy a CPI release with a stemcell", func() {
		contents := `---
name: test-release
cloud_provider:
  properties:
    cpi:
      warden:
        connect_network: unix
        connect_address: /var/vcap/data/warden/warden.sock
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
	})
})

type acceptanceEnvironment interface {
	Path(string) string
	Copy(string, string) error
	WriteContentString(string, string) error
	Cleanup() error
}

type vagrantTestEnvironment struct {
	testRoot   string
	fileSystem boshsys.FileSystem
}

func newVagrantTestEnvironment(fileSystem boshsys.FileSystem) *vagrantTestEnvironment {
	fileSystem.MkdirAll("./../tmp", os.FileMode(0775))
	return &vagrantTestEnvironment{
		testRoot:   "./../tmp",
		fileSystem: fileSystem,
	}
}

func (e *vagrantTestEnvironment) Path(name string) string {
	return path.Join("/", "vagrant", name)
}

func (e *vagrantTestEnvironment) Copy(destName, srcPath string) error {
	if srcPath == "" {
		return fmt.Errorf("Cannot use an empty file for `%s'", destName)
	}
	// the testRoot dir will be synced or mounted inside the vagrant box
	// by Vagrant itself (see Vagrantfile).
	return e.fileSystem.CopyFile(srcPath, path.Join(e.testRoot, destName))
}

func (e *vagrantTestEnvironment) WriteContentString(destName, contents string) error {
	return e.fileSystem.WriteFileString(path.Join(e.testRoot, destName), contents)
}

func (e *vagrantTestEnvironment) Cleanup() error {
	return e.fileSystem.RemoveAll(e.testRoot)
}

type acceptanceCmdRunner interface {
	StartSession() error
	EndSession()
	RunCommand(...string) (string, string, int, error)
}

type vagrantSSHCmdRunner struct {
	runner boshsys.CmdRunner
	logger boshlog.Logger
}

func NewVagrantSSHCmdRunner(logger boshlog.Logger) vagrantSSHCmdRunner {
	return vagrantSSHCmdRunner{
		runner: boshsys.NewExecCmdRunner(logger),
		logger: logger,
	}
}

func (r vagrantSSHCmdRunner) StartSession() error {
	r.runner.RunCommand("vagrant", "destroy", "-f")

	_, _, exitCode, err := r.runner.RunCommand("vagrant", "up")
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return errors.New("Failed to bring up vagrant VM")
	}

	return nil
}

func (r vagrantSSHCmdRunner) EndSession() {
	r.runner.RunCommand("vagrant", "destroy", "-f")
}

func (r vagrantSSHCmdRunner) RunCommand(args ...string) (string, string, int, error) {
	argsWithEnv := append([]string{"TMPDIR=/home/vcap"}, args...)
	vcapArgs := []string{
		"sudo", "su", "vcap", "-c",
		fmt.Sprintf("'%s'", strings.Join(argsWithEnv, " ")),
	}
	return r.runner.RunCommand("vagrant", "ssh", "-c", strings.Join(vcapArgs, " "))
}
