package acceptance_test

import (
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
		sshCmdRunner CmdRunner
		testEnv      Environment
		config       *Config

		microSSH      MicroSSH
		microUsername = "vcap"
		microPassword = "sshpassword"
		microIP       = "10.244.0.42"
	)

	BeforeSuite(func() {
		logger := boshlog.NewLogger(boshlog.LevelDebug)
		fileSystem = boshsys.NewOsFileSystem(logger)

		var err error
		config, err = NewConfig(fileSystem)
		Expect(err).NotTo(HaveOccurred())

		err = config.Validate()
		Expect(err).NotTo(HaveOccurred())

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

		microSSH = NewMicroSSH(
			config.VMUsername,
			config.VMIP,
			config.PrivateKeyPath,
			microUsername,
			microIP,
			microPassword,
			fileSystem,
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

	AfterEach(func() {
		_, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "delete", testEnv.Path("cpiRelease"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
	})

	var setDeploymentManifest = func(manifestPath string) {
		manifestContents, err := ioutil.ReadFile(manifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)
	}

	It("can set deployment, deploy, update, and delete", func() {
		setDeploymentManifest("./manifest.yml")

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
		Expect(stdout).To(ContainSubstring("Creating VM for instance 'bosh/0' from stemcell"))
		Expect(stdout).To(ContainSubstring("Waiting for the agent on VM"))
		Expect(stdout).To(ContainSubstring("Creating disk"))
		Expect(stdout).To(ContainSubstring("Attaching disk"))
		Expect(stdout).To(ContainSubstring("Starting instance 'bosh/0'"))
		Expect(stdout).To(ContainSubstring("Waiting for instance 'bosh/0' to be running"))
		Expect(stdout).To(ContainSubstring("Done deploying"))
	})

	Context("when microbosh has been previously deployed", func() {
		BeforeEach(func() {
			setDeploymentManifest("./manifest.yml")

			_, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deployment", testEnv.Path("manifest"))
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			_, _, exitCode, err = sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
		})

		It("sets the ssh password", func() {
			stdout, _, exitCode, err := microSSH.RunCommand("echo ssh-succeeded")
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(stdout).To(ContainSubstring("ssh-succeeded"))
		})

		It("when there are no changes, it skips deploy", func() {
			stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(stdout).To(ContainSubstring("No deployment, stemcell or cpi release changes. Skipping deploy."))
			Expect(stdout).ToNot(ContainSubstring("Started installing CPI jobs"))
			Expect(stdout).ToNot(ContainSubstring("Started deploying"))
		})

		It("when updating with property changes, it deletes the old VM", func() {
			setDeploymentManifest("./modified_manifest.yml")

			stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Stopping jobs on instance 'unknown/0'"))
			Expect(stdout).To(ContainSubstring("Unmounting disk"))

			Expect(stdout).ToNot(ContainSubstring("Creating disk"))
		})

		It("when updating with disk size changed, it migrates the disk", func() {
			setDeploymentManifest("./modified_disk_manifest.yml")

			stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Stopping jobs on instance 'unknown/0'"))
			Expect(stdout).To(ContainSubstring("Unmounting disk"))

			Expect(stdout).To(ContainSubstring("Creating disk"))
			Expect(stdout).To(ContainSubstring("Migrating disk"))
			Expect(stdout).To(ContainSubstring("Deleting disk"))
		})

		It("can delete all vms, disk, and stemcells", func() {
			stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "delete", testEnv.Path("cpiRelease"))
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(stdout).To(ContainSubstring("Stopping jobs on instance"))
			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Deleting disk"))
			Expect(stdout).To(ContainSubstring("Deleting stemcell"))
			Expect(stdout).To(ContainSubstring("Done deleting deployment"))
		})

		Context("when the agent is unresponsive", func() {
			BeforeEach(func() {
				_, _, exitCode, err := microSSH.RunCommandWithSudo("sv -w 14 force-shutdown agent")
				Expect(err).ToNot(HaveOccurred())
				Expect(exitCode).To(Equal(0))
			})

			It("re-deploys if the agent is unresponsive", func() {
				setDeploymentManifest("./modified_manifest.yml")

				stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
				Expect(err).ToNot(HaveOccurred())
				Expect(exitCode).To(Equal(0))

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'... failed."))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Creating VM for instance 'bosh/0' from stemcell"))
				Expect(stdout).To(ContainSubstring("Done deploying"))
			})

			It("deletes if the agent is unresponsive", func() {
				stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "delete", testEnv.Path("cpiRelease"))
				Expect(err).ToNot(HaveOccurred())
				Expect(exitCode).To(Equal(0))

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'... failed."))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Deleting disk"))
				Expect(stdout).To(ContainSubstring("Deleting stemcell"))
				Expect(stdout).To(ContainSubstring("Done deleting deployment"))
			})
		})
	})

	It("deploys without registry and ssh tunnel", func() {
		setDeploymentManifest("./manifest_without_registry.yml")

		_, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deployment", testEnv.Path("manifest"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		Expect(stdout).To(ContainSubstring("Done deploying"))
	})
})
