package acceptance_test

import (
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"

	. "github.com/cloudfoundry/bosh-micro-cli/acceptance"
)

var _ = Describe("bosh-micro", func() {
	var (
		fileSystem   boshsys.FileSystem
		cmdRunner    boshsys.CmdRunner
		sshCmdRunner CmdRunner
		testEnv      Environment
		config       *Config
	)

	BeforeSuite(func() {
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		fileSystem = boshsys.NewOsFileSystem(logger)

		var err error
		config, err = NewConfig(fileSystem)
		Expect(err).NotTo(HaveOccurred())

		err = config.Validate()
		Expect(err).NotTo(HaveOccurred())

		cmdRunner = boshsys.NewExecCmdRunner(logger)

		testEnv = NewRemoteTestEnvironment(
			config.VMUsername,
			config.VMIP,
			config.PrivateKeyPath,
			fileSystem,
			logger,
		)

		sshCmdRunner = NewSSHCmdRunner(
			config.VMUsername,
			config.VMIP,
			config.PrivateKeyPath,
			logger,
		)

		err = bmtestutils.BuildExecutableForArch("linux-amd64")
		Expect(err).NotTo(HaveOccurred())

		boshMicroPath := "./../out/bosh-micro"
		Expect(fileSystem.FileExists(boshMicroPath)).To(BeTrue())
		err = testEnv.Copy("bosh-micro", boshMicroPath)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("stemcell", config.StemcellPath, config.StemcellURL)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("cpiRelease", config.CpiReleasePath, config.CpiReleaseURL)
		Expect(err).NotTo(HaveOccurred())
	})

	ItSetsSSHPassword := func(username, password, hostname string) {
		sshConfigFile, err := fileSystem.TempFile("ssh-config")
		Expect(err).ToNot(HaveOccurred())
		defer fileSystem.RemoveAll(sshConfigFile.Name())

		sshConfigTemplate := `
Host vagrant-vm
  HostName %s
  User %s
  Port 22
  StrictHostKeyChecking no
  IdentityFile %s
Host warden-vm
  Hostname %s
  User %s
  StrictHostKeyChecking no
  ProxyCommand ssh -F %s vagrant-vm netcat -w 120 %%h %%p
`
		sshConfig := fmt.Sprintf(
			sshConfigTemplate,
			config.VMIP,
			config.VMUsername,
			config.PrivateKeyPath,
			hostname,
			username,
			sshConfigFile.Name(),
		)

		fileSystem.WriteFileString(sshConfigFile.Name(), sshConfig)
		stdout, _, exitCode, err := cmdRunner.RunCommand(
			"sshpass",
			"-p"+password,
			"ssh",
			"warden-vm",
			"-F",
			sshConfigFile.Name(),
			"echo ssh-succeeded",
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		Expect(stdout).To(ContainSubstring("ssh-succeeded"))
	}

	It("is able to deploy a CPI release with a stemcell", func() {
		manifestPath := "./manifest.yml"
		manifestContents, err := ioutil.ReadFile(manifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)

		_, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deployment", testEnv.Path("manifest"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		Expect(stdout).To(ContainSubstring("Uploading stemcell"))
		Expect(stdout).To(ContainSubstring("Creating VM from"))
		Expect(stdout).To(ContainSubstring("Waiting for the agent"))
		Expect(stdout).To(ContainSubstring("Applying micro BOSH spec"))
		Expect(stdout).To(ContainSubstring("Starting agent services"))
		Expect(stdout).To(ContainSubstring("Waiting for the director"))
		Expect(stdout).To(ContainSubstring("Creating disk"))

		ItSetsSSHPassword("vcap", "sshpassword", "10.244.0.42")
	})
})
