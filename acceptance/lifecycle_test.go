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

	ItSkipsDeployIfNoChanges := func() {
		stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		Expect(stdout).To(ContainSubstring("No deployment, stemcell or cpi release changes. Skipping deploy."))
		Expect(stdout).ToNot(ContainSubstring("Started installing CPI jobs"))
		Expect(stdout).ToNot(ContainSubstring("Started deploying"))
	}

	ItDeletesVMOnUpdate := func() {
		manifestPath := "./modified_manifest.yml"
		manifestContents, err := ioutil.ReadFile(manifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)

		stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		Expect(stdout).To(ContainSubstring("Deleting VM"))
		Expect(stdout).To(ContainSubstring("Stopping 'bosh'"))
		Expect(stdout).To(ContainSubstring("Unmounting disk"))

		Expect(stdout).ToNot(ContainSubstring("Creating disk"))
	}

	ItMigratesDisk := func() {
		manifestPath := "./modified_disk_manifest.yml"
		manifestContents, err := ioutil.ReadFile(manifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)

		stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		Expect(stdout).To(ContainSubstring("Deleting VM"))
		Expect(stdout).To(ContainSubstring("Stopping 'bosh'"))
		Expect(stdout).To(ContainSubstring("Unmounting disk"))

		Expect(stdout).To(ContainSubstring("Creating disk"))
		Expect(stdout).To(ContainSubstring("Migrating disk"))
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

		Expect(stdout).To(ContainSubstring("Started validating"))
		Expect(stdout).To(ContainSubstring("Validating deployment manifest"))
		Expect(stdout).To(ContainSubstring("Validating cpi release"))
		Expect(stdout).To(ContainSubstring("Validating stemcell"))
		Expect(stdout).To(ContainSubstring("Done validating"))

		Expect(stdout).To(ContainSubstring("Started compiling packages"))
		Expect(stdout).To(ContainSubstring("Done compiling packages"))

		Expect(stdout).To(ContainSubstring("Started installing CPI jobs"))
		Expect(stdout).To(ContainSubstring("Done installing CPI jobs"))

		Expect(stdout).To(ContainSubstring("Started uploading stemcell"))
		Expect(stdout).To(ContainSubstring("Done uploading stemcell"))

		Expect(stdout).To(ContainSubstring("Started deploying"))
		Expect(stdout).To(ContainSubstring("Creating VM from stemcell"))
		Expect(stdout).To(ContainSubstring("Waiting for the agent"))
		Expect(stdout).To(ContainSubstring("Creating disk"))
		Expect(stdout).To(ContainSubstring("Attaching disk"))
		Expect(stdout).To(ContainSubstring("Starting 'bosh'"))
		Expect(stdout).To(ContainSubstring("Waiting for 'bosh'"))
		Expect(stdout).To(ContainSubstring("Done deploying"))

		ItSetsSSHPassword("vcap", "sshpassword", "10.244.0.42")

		ItSkipsDeployIfNoChanges()

		ItDeletesVMOnUpdate()

		ItMigratesDisk()
	})
})
