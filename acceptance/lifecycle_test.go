package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"os"
	"bytes"

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
		quietCmdEnv  map[string]string
		testEnv      Environment
		config       *Config

		microSSH      MicroSSH
		microUsername = "vcap"
		microPassword = "sshpassword" // encrypted value must be in the manifest: resource_pool.env.bosh.password
		microIP       = "10.244.0.42"
	)

	var readLogFile = func(logPath string) (stdout string) {
		stdout, _, exitCode, err := sshCmdRunner.RunCommand(cmdEnv, "cat", logPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return stdout
	}

	var deleteLogFile = func(logPath string) {
		_, _, exitCode, err := sshCmdRunner.RunCommand(cmdEnv, "rm", "-f", logPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
	}

	var flushLog = func(logPath string) {
		//TODO: stream command stdout/stderr to GinkgoWriter
		logString := readLogFile(logPath)
		_, err := GinkgoWriter.Write([]byte(logString))
		Expect(err).ToNot(HaveOccurred())

		// only delete after successfully writing to GinkgoWriter
		deleteLogFile(logPath)
	}

	BeforeSuite(func() {
		// writing to GinkgoWriter prints on test failure or when using verbose mode (-v)
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, GinkgoWriter, GinkgoWriter)
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
			"TMPDIR":               testEnv.Home(),
			"BOSH_MICRO_LOG_LEVEL": "DEBUG",
			"BOSH_MICRO_LOG_PATH":  testEnv.Path("bosh-micro-cli.log"),
		}
		quietCmdEnv = map[string]string{
			"TMPDIR":               testEnv.Home(),
			"BOSH_MICRO_LOG_LEVEL": "ERROR",
			"BOSH_MICRO_LOG_PATH":  testEnv.Path("bosh-micro-cli-cleanup.log"),
		}

		// clean up from previous failed tests
		deleteLogFile(cmdEnv["BOSH_MICRO_LOG_PATH"])
		deleteLogFile(quietCmdEnv["BOSH_MICRO_LOG_PATH"])

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
		err = testEnv.DownloadOrCopy("stemcell.tgz", config.StemcellPath, config.StemcellURL)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("cpi-release.tgz", config.CpiReleasePath, config.CpiReleaseURL)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("bosh-release.tgz", config.BoshReleasePath, config.BoshReleaseURL)
		Expect(err).NotTo(HaveOccurred())
	})

	// updateDeploymentManifest copies a source manifest from assets to <workspace>/manifest
	var updateDeploymentManifest = func(sourceManifestPath string) {
		manifestContents, err := ioutil.ReadFile(sourceManifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)
	}

	var setDeployment = func(manifestPath string) (stdout string) {
		os.Stdout.WriteString("\n---DEPLOYMENT---\n")
		outBuffer := bytes.NewBufferString("")
		multiWriter := NewMultiWriter(outBuffer, os.Stdout)
		_, _, exitCode, err := sshCmdRunner.RunStreamingCommand(multiWriter, cmdEnv, testEnv.Path("bosh-micro"), "deployment", manifestPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return outBuffer.String()
	}

	var deploy = func() (stdout string) {
		os.Stdout.WriteString("\n---DEPLOY---\n")
		outBuffer := bytes.NewBufferString("")
		multiWriter := NewMultiWriter(outBuffer, os.Stdout)
		_, _, exitCode, err := sshCmdRunner.RunStreamingCommand(multiWriter, cmdEnv, testEnv.Path("bosh-micro"), "deploy", testEnv.Path("stemcell.tgz"), testEnv.Path("cpi-release.tgz"), testEnv.Path("bosh-release.tgz"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return outBuffer.String()
	}

	var deleteDeployment = func() (stdout string) {
		os.Stdout.WriteString("\n---DELETE---\n")
		outBuffer := bytes.NewBufferString("")
		multiWriter := NewMultiWriter(outBuffer, os.Stdout)
		_, _, exitCode, err := sshCmdRunner.RunStreamingCommand(multiWriter, cmdEnv, testEnv.Path("bosh-micro"), "delete", testEnv.Path("cpi-release.tgz"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return outBuffer.String()
	}

	AfterEach(func() {
		flushLog(cmdEnv["BOSH_MICRO_LOG_PATH"])

		// quietly delete the deployment
		_, _, exitCode, err := sshCmdRunner.RunCommand(quietCmdEnv, testEnv.Path("bosh-micro"), "delete", testEnv.Path("cpi-release.tgz"))
		if exitCode != 0 || err != nil {
			// only flush the delete log if the delete failed
			flushLog(quietCmdEnv["BOSH_MICRO_LOG_PATH"])
		}
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
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
		err = tempUserConfigFile.Close()
		Expect(err).ToNot(HaveOccurred())
		err = fileSystem.WriteFileString(tempUserConfigFile.Name(), stdout)
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
		Expect(stdout).To(ContainSubstring(fmt.Sprintf("Deployment manifest set to '%s'", manifestPath)))

		Expect(parseUserConfig()).To(Equal(bmconfig.UserConfig{
			DeploymentManifestPath: manifestPath,
		}))
	})

	var findStage = func(outputLines []string, stageName string, zeroIndex int) (steps []string, stopIndex int) {
		startLine := fmt.Sprintf("Started %s", stageName)
		startIndex := -1
		for i, line := range outputLines[zeroIndex:] {
			if line == startLine {
				startIndex = zeroIndex + i
				break
			}
		}
		if startIndex < 0 {
			Fail("Failed to find stage start: " + stageName)
		}

		stopLine := fmt.Sprintf("Done %s", stageName)
		stopIndex = -1
		for i, line := range outputLines[startIndex:] {
			if line == stopLine {
				stopIndex = startIndex + i
				break
			}
		}
		if stopIndex < 0 {
			Fail("Failed to find stage stop: " + stageName)
		}

		return outputLines[startIndex+1 : stopIndex], stopIndex
	}

	It("can deploy", func() {
		updateDeploymentManifest("./assets/manifest.yml")

		setDeployment(testEnv.Path("manifest"))

		stdout := deploy()

		outputLines := strings.Split(stdout, "\n")

		donePattern := "\\.\\.\\. done\\. \\(\\d{2}:\\d{2}:\\d{2}\\)$"

		doneIndex := 0

		validatingSteps, doneIndex := findStage(outputLines, "validating", doneIndex)
		Expect(validatingSteps[0]).To(MatchRegexp("^Started validating > Validating stemcell" + donePattern))
		Expect(validatingSteps[1]).To(MatchRegexp("^Started validating > Validating releases" + donePattern))
		Expect(validatingSteps[2]).To(MatchRegexp("^Started validating > Validating deployment manifest" + donePattern))
		Expect(validatingSteps[3]).To(MatchRegexp("^Started validating > Validating cpi release" + donePattern))
		Expect(validatingSteps).To(HaveLen(4))

		compilingSteps, doneIndex := findStage(outputLines, "compiling packages", doneIndex+1)
		for _, line := range compilingSteps {
			Expect(line).To(MatchRegexp("^Started compiling packages > .*/.*" + donePattern))
		}
		Expect(len(compilingSteps)).To(BeNumerically(">", 0))

		installingSteps, doneIndex := findStage(outputLines, "installing CPI jobs", doneIndex+1)
		Expect(installingSteps[0]).To(MatchRegexp("^Started installing CPI jobs > cpi" + donePattern))
		Expect(installingSteps).To(HaveLen(1))

		uploadingSteps, doneIndex := findStage(outputLines, "uploading stemcell", doneIndex+1)
		Expect(uploadingSteps[0]).To(MatchRegexp("^Started uploading stemcell > Uploading" + donePattern))
		Expect(uploadingSteps).To(HaveLen(1))

		deployingSteps, doneIndex := findStage(outputLines, "deploying", doneIndex+1)
		numDeployingSteps := len(deployingSteps)
		Expect(deployingSteps[0]).To(MatchRegexp("^Started deploying > Creating VM for instance 'bosh/0' from stemcell '.*'" + donePattern))
		Expect(deployingSteps[1]).To(MatchRegexp("^Started deploying > Waiting for the agent on VM '.*' to be ready" + donePattern))
		Expect(deployingSteps[2]).To(MatchRegexp("^Started deploying > Creating disk" + donePattern))
		Expect(deployingSteps[3]).To(MatchRegexp("^Started deploying > Attaching disk '.*' to VM '.*'" + donePattern))
		for _, line := range deployingSteps[4:numDeployingSteps-3] {
			Expect(line).To(MatchRegexp("^Started deploying > Compiling package '.*/.*'" + donePattern))
		}
		Expect(deployingSteps[numDeployingSteps-3]).To(MatchRegexp("^Started deploying > Rendering job templates" + donePattern))
		Expect(deployingSteps[numDeployingSteps-2]).To(MatchRegexp("^Started deploying > Updating instance 'bosh/0'" + donePattern))
		Expect(deployingSteps[numDeployingSteps-1]).To(MatchRegexp("^Started deploying > Waiting for instance 'bosh/0' to be running" + donePattern))
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

			It("re-deploying deletes the vm", func() {
				updateDeploymentManifest("./assets/modified_manifest.yml")

				stdout := deploy()

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'\\.\\.\\. failed."))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Creating VM for instance 'bosh/0' from stemcell"))
				Expect(stdout).To(ContainSubstring("Done deploying"))
			})

			It("delete deletes the vm", func() {
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
