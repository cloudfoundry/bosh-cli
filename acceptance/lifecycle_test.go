package acceptance_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/acceptance"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

const (
	stageTimePattern     = "\\(\\d{2}:\\d{2}:\\d{2}\\)"
	stageFinishedPattern = "\\.\\.\\. Finished " + stageTimePattern + "$"
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

	var expectDeployToError = func() (stdout string) {
		os.Stdout.WriteString("\n---DEPLOY---\n")
		outBuffer := bytes.NewBufferString("")
		multiWriter := NewMultiWriter(outBuffer, os.Stdout)
		_, _, exitCode, err := sshCmdRunner.RunStreamingCommand(multiWriter, cmdEnv, testEnv.Path("bosh-micro"), "deploy", testEnv.Path("stemcell.tgz"), testEnv.Path("cpi-release.tgz"), testEnv.Path("bosh-release.tgz"))
		Expect(err).To(HaveOccurred())
		Expect(exitCode).To(Equal(1))
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

		stopLinePattern := fmt.Sprintf("^Finished %s %s$", stageName, stageTimePattern)
		stopLineRegex, err := regexp.Compile(stopLinePattern)
		Expect(err).ToNot(HaveOccurred())

		stopIndex = -1
		for i, line := range outputLines[startIndex:] {
			if stopLineRegex.MatchString(line) {
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

		doneIndex := 0

		validatingSteps, doneIndex := findStage(outputLines, "validating", doneIndex)
		Expect(validatingSteps[0]).To(MatchRegexp("^  Validating stemcell" + stageFinishedPattern))
		Expect(validatingSteps[1]).To(MatchRegexp("^  Validating releases" + stageFinishedPattern))
		Expect(validatingSteps[2]).To(MatchRegexp("^  Validating deployment manifest" + stageFinishedPattern))
		Expect(validatingSteps[3]).To(MatchRegexp("^  Validating cpi release" + stageFinishedPattern))
		Expect(validatingSteps).To(HaveLen(4))

		installingSteps, doneIndex := findStage(outputLines, "installing CPI", doneIndex+1)
		numInstallingSteps := len(installingSteps)
		for _, line := range installingSteps[:numInstallingSteps-3] {
			Expect(line).To(MatchRegexp("^  Compiling package '.*/.*'" + stageFinishedPattern))
		}
		Expect(installingSteps[numInstallingSteps-3]).To(MatchRegexp("^  Rendering job templates" + stageFinishedPattern))
		Expect(installingSteps[numInstallingSteps-2]).To(MatchRegexp("^  Installing packages" + stageFinishedPattern))
		Expect(installingSteps[numInstallingSteps-1]).To(MatchRegexp("^  Installing job 'cpi'" + stageFinishedPattern))

		Expect(outputLines[doneIndex+2]).To(MatchRegexp("^Starting registry" + stageFinishedPattern))
		Expect(outputLines[doneIndex+3]).To(MatchRegexp("^Uploading stemcell '.*/.*'" + stageFinishedPattern))

		deployingSteps, doneIndex := findStage(outputLines, "deploying", doneIndex+1)
		numDeployingSteps := len(deployingSteps)
		Expect(deployingSteps[0]).To(MatchRegexp("^  Creating VM for instance 'bosh/0' from stemcell '.*'" + stageFinishedPattern))
		Expect(deployingSteps[1]).To(MatchRegexp("^  Waiting for the agent on VM '.*' to be ready" + stageFinishedPattern))
		Expect(deployingSteps[2]).To(MatchRegexp("^  Creating disk" + stageFinishedPattern))
		Expect(deployingSteps[3]).To(MatchRegexp("^  Attaching disk '.*' to VM '.*'" + stageFinishedPattern))
		for _, line := range deployingSteps[4 : numDeployingSteps-3] {
			Expect(line).To(MatchRegexp("^  Compiling package '.*/.*'" + stageFinishedPattern))
		}
		Expect(deployingSteps[numDeployingSteps-3]).To(MatchRegexp("^  Rendering job templates" + stageFinishedPattern))
		Expect(deployingSteps[numDeployingSteps-2]).To(MatchRegexp("^  Updating instance 'bosh/0'" + stageFinishedPattern))
		Expect(deployingSteps[numDeployingSteps-1]).To(MatchRegexp("^  Waiting for instance 'bosh/0' to be running" + stageFinishedPattern))
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
			Expect(stdout).To(ContainSubstring("Finished deleting deployment"))
		})

		Context("when the agent is unresponsive", func() {
			BeforeEach(func() {
				_, _, exitCode, err := microSSH.RunCommandWithSudo("sv -w 14 force-shutdown agent")
				if exitCode == 1 {
					// If timeout was reached, KILL signal was sent before exiting.
					// Retry to wait another 14s for exit.
					_, _, exitCode, err = microSSH.RunCommandWithSudo("sv -w 14 force-shutdown agent")
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(exitCode).To(Equal(0))
			})

			It("re-deploying deletes the vm", func() {
				updateDeploymentManifest("./assets/modified_manifest.yml")

				stdout := deploy()

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'\\.\\.\\. Failed " + stageTimePattern))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Creating VM for instance 'bosh/0' from stemcell"))
				Expect(stdout).To(ContainSubstring("Finished deploying"))
			})

			It("delete deletes the vm", func() {
				stdout := deleteDeployment()

				Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'\\.\\.\\. Failed " + stageTimePattern))
				Expect(stdout).To(ContainSubstring("Deleting VM"))
				Expect(stdout).To(ContainSubstring("Deleting disk"))
				Expect(stdout).To(ContainSubstring("Deleting stemcell"))
				Expect(stdout).To(ContainSubstring("Finished deleting deployment"))
			})
		})
	})

	It("deploys & deletes without registry and ssh tunnel", func() {
		updateDeploymentManifest("./assets/manifest_without_registry.yml")

		setDeployment(testEnv.Path("manifest"))

		stdout := deploy()
		Expect(stdout).To(ContainSubstring("Finished deploying"))

		stdout = deleteDeployment()
		Expect(stdout).To(ContainSubstring("Finished deleting deployment"))
	})

	It("prints multiple validation errors at the same time", func() {
		updateDeploymentManifest("./assets/invalid_manifest.yml")

		setDeployment(testEnv.Path("manifest"))

		stdout := expectDeployToError()

		Expect(stdout).To(ContainSubstring("Validating deployment manifest... Failed"))
		Expect(stdout).To(ContainSubstring("Failed validating"))

		Expect(stdout).To(ContainSubstring(`
Command 'deploy' failed:
  Validating deployment manifest:
    jobs[0].templates[0].release must refer to an available release:
      Release 'unknown-release' is not available
    jobs[0].templates[5].release must refer to an available release:
      Release 'unknown-release' is not available`))
	})
})
