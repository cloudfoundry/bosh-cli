package acceptance_test

import (
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	. "github.com/cloudfoundry/bosh-micro-cli/acceptance"
)

var _ = Describe("bosh-micro", func() {
	var (
		logger       boshlog.Logger
		fileSystem   boshsys.FileSystem
		sshCmdRunner CmdRunner
		cmdEnv       map[string]string
		testEnv      Environment
		config       *Config

		microSSH      MicroSSH
		microUsername = "vcap"
		microPassword = "sshpassword"
		microIP       = "10.244.0.42"
	)

	BeforeSuite(func() {
		logger = boshlog.NewLogger(boshlog.LevelDebug)
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
		cmdEnv = map[string]string{
			"TMPDIR":         fmt.Sprintf("/home/%s", config.VMUsername),
			"BOSH_MICRO_LOG": "DEBUG",
		}

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

	// updateDeploymentManifest copies a source manifest from assets to <workspace>/manifest
	var updateDeploymentManifest = func(sourceManifestPath string) {
		manifestContents, err := ioutil.ReadFile(sourceManifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)
	}

	var setDeployment = func(manifestPath string) (stdout string) {
		stdout, _, exitCode, err := sshCmdRunner.RunCommand(cmdEnv, testEnv.Path("bosh-micro"), "deployment", manifestPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return stdout
	}

	var deploy = func() (stdout string) {
		stdout, _, exitCode, err := sshCmdRunner.RunCommand(cmdEnv, testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return stdout
	}

	var deleteDeployment = func() (stdout string) {
		stdout, _, exitCode, err := sshCmdRunner.RunCommand(cmdEnv, testEnv.Path("bosh-micro"), "delete", testEnv.Path("cpiRelease"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return stdout
	}

	AfterEach(func() {
		deleteDeployment()
	})

	// parseUserConfig reads & parses the remote bosh-micro user config
	// This would be a lot cleaner if there were a RemoteFileSystem that used SSH.
	var parseUserConfig = func() bmconfig.UserConfig {
		userConfigPath := testEnv.Path(".bosh_micro.json")
		stdout, _, exitCode, err := sshCmdRunner.RunCommand(cmdEnv, "cat", userConfigPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		tempUserConfigFile, err := fileSystem.TempFile("bosh-micro-user-config")
		Expect(err).ToNot(HaveOccurred())
		_, err = tempUserConfigFile.WriteString(stdout)
		Expect(err).ToNot(HaveOccurred())
		defer fileSystem.RemoveAll(tempUserConfigFile.Name())

		userConfigService := bmconfig.NewFileSystemUserConfigService(tempUserConfigFile.Name(), fileSystem, logger)
		userConfig, err := userConfigService.Load()
		Expect(err).ToNot(HaveOccurred())

		return userConfig
	}

	It("can set deployment", func() {
		updateDeploymentManifest("./assets/manifest.yml")

		manifestPath := testEnv.Path("manifest")

		stdout := setDeployment(manifestPath)
		Expect(stdout).To(ContainSubstring(fmt.Sprintf("Deployment set to `%s'", manifestPath)))

		Expect(parseUserConfig()).To(Equal(bmconfig.UserConfig{
			DeploymentFile: manifestPath,
		}))
	})

	It("can deploy", func() {
		updateDeploymentManifest("./assets/manifest.yml")

		setDeployment(testEnv.Path("manifest"))

		stdout := deploy()

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
			updateDeploymentManifest("./assets/manifest.yml")

			setDeployment(testEnv.Path("manifest"))

			deploy()
		})

		It("sets the ssh password", func() {
			stdout, _, exitCode, err := microSSH.RunCommand("echo ssh-succeeded")
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(stdout).To(ContainSubstring("ssh-succeeded"))
		})

		It("when there are no changes, it skips deploy", func() {
			stdout := deploy()

			Expect(stdout).To(ContainSubstring("No deployment, stemcell or cpi release changes. Skipping deploy."))
			Expect(stdout).ToNot(ContainSubstring("Started installing CPI jobs"))
			Expect(stdout).ToNot(ContainSubstring("Started deploying"))
		})

		It("when updating with property changes, it deletes the old VM", func() {
			updateDeploymentManifest("./assets/modified_manifest.yml")

			stdout := deploy()

			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Stopping jobs on instance 'unknown/0'"))
			Expect(stdout).To(ContainSubstring("Unmounting disk"))

			Expect(stdout).ToNot(ContainSubstring("Creating disk"))
		})

		It("when updating with disk size changed, it migrates the disk", func() {
			updateDeploymentManifest("./assets/modified_disk_manifest.yml")

			stdout := deploy()

			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Stopping jobs on instance 'unknown/0'"))
			Expect(stdout).To(ContainSubstring("Unmounting disk"))

			Expect(stdout).To(ContainSubstring("Creating disk"))
			Expect(stdout).To(ContainSubstring("Migrating disk"))
			Expect(stdout).To(ContainSubstring("Deleting disk"))
		})

		It("can delete all vms, disk, and stemcells", func() {
			stdout := deleteDeployment()

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
				updateDeploymentManifest("./assets/modified_manifest.yml")

				stdout := deploy()

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'\\.\\.\\. failed."))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Creating VM for instance 'bosh/0' from stemcell"))
				Expect(stdout).To(ContainSubstring("Done deploying"))
			})

			It("deletes if the agent is unresponsive", func() {
				stdout := deleteDeployment()

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'\\.\\.\\. failed."))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Deleting disk"))
				Expect(stdout).To(ContainSubstring("Deleting stemcell"))
				Expect(stdout).To(ContainSubstring("Done deleting deployment"))
			})
		})
	})

	It("deploys & deletes without registry and ssh tunnel", func() {
		updateDeploymentManifest("./assets/manifest_without_registry.yml")

		setDeployment(testEnv.Path("manifest"))

		stdout := deploy()
		Expect(stdout).To(ContainSubstring("Done deploying"))

		stdout = deleteDeployment()
		Expect(stdout).To(ContainSubstring("Done deleting deployment"))
	})
})
