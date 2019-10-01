package acceptance_test

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"

	. "github.com/cloudfoundry/bosh-cli/acceptance"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	bitestutils "github.com/cloudfoundry/bosh-cli/testutils"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const (
	stageTimePattern                   = "\\(\\d{2}:\\d{2}:\\d{2}\\)"
	stageFinishedPattern               = "\\.\\.\\. Finished " + stageTimePattern + "$"
	stageCompiledPackageSkippedPattern = "\\.\\.\\. Skipped \\[Package already compiled\\] " + stageTimePattern + "$"
)

var _ = Describe("bosh", func() {
	var (
		logger          boshlog.Logger
		fileSystem      boshsys.FileSystem
		cmdRunner       CmdRunner
		cmdEnv          map[string]string
		quietCmdEnv     map[string]string
		testEnv         Environment
		config          *Config
		extraDeployArgs []string

		instanceSSH      InstanceSSH
		instanceUsername = "vcap"
		instancePassword = "sshpassword" // encrypted value must be in the manifest: resource_pool.env.bosh.password
		instanceIP       = "10.244.0.42"
	)

	var readLogFile = func(logPath string) (stdout string) {
		stdout, _, exitCode, err := cmdRunner.RunCommand(cmdEnv, "cat", logPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		return stdout
	}

	var deleteLogFile = func(logPath string) {
		_, _, exitCode, err := cmdRunner.RunCommand(cmdEnv, "rm", "-f", logPath)
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

	type manifestContext struct {
		CPIReleaseURL            string
		StemcellURL              string
		DummyCompiledReleasePath string
	}

	var prepareDeploymentManifest = func(context manifestContext, sourceManifestPath string) []byte {
		context.CPIReleaseURL = "file://" + testEnv.Path("cpi-release.tgz")
		context.StemcellURL = "file://" + testEnv.Path("stemcell.tgz")

		buffer := &bytes.Buffer{}
		t := template.Must(template.ParseFiles(sourceManifestPath))
		err := t.Execute(buffer, context)
		Expect(err).ToNot(HaveOccurred())

		return buffer.Bytes()
	}

	var updateCompiledReleaseDeploymentManifest = func(sourceManifestPath string) {
		context := manifestContext{
			DummyCompiledReleasePath: testEnv.Path("sample-release-compiled.tgz"),
		}

		buffer := prepareDeploymentManifest(context, sourceManifestPath)
		err := testEnv.WriteContent("test-manifest.yml", buffer)
		Expect(err).NotTo(HaveOccurred())
	}

	var setupSshKey = func(varsFile string) {
		stdout := &bytes.Buffer{}
		multiWriter := io.MultiWriter(stdout, GinkgoWriter)

		args := append([]string{testEnv.Path("bosh"), "int", varsFile}, "--path /ssh_tunnel/private_key")

		_, _, _, err := cmdRunner.RunStreamingCommand(multiWriter, cmdEnv, args...)
		Expect(err).ToNot(HaveOccurred())

		Expect(fileSystem.WriteFile("/tmp/test_private_key", stdout.Bytes())).To(Succeed())
		Expect(fileSystem.Chmod("/tmp/test_private_key", os.FileMode(0400))).To(Succeed())
	}

	var deploy = func(manifestFile string, args ...string) string {
		fmt.Fprintf(GinkgoWriter, "\n--- DEPLOY ---\n")

		stdout := &bytes.Buffer{}
		multiWriter := io.MultiWriter(stdout, GinkgoWriter)

		args = append([]string{testEnv.Path("bosh"), "create-env", "--tty", testEnv.Path(manifestFile)}, args...)

		_, _, exitCode, err := cmdRunner.RunStreamingCommand(multiWriter, cmdEnv, args...)
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		setupSshKey(testEnv.Path("vars.yml"))

		return stdout.String()
	}

	var deleteDeployment = func(manifest string, extraDeployArgs ...string) string {
		fmt.Fprintf(GinkgoWriter, "\n--- DELETE DEPLOYMENT ---\n")

		stdout := &bytes.Buffer{}
		multiWriter := io.MultiWriter(stdout, GinkgoWriter)

		args := append([]string{testEnv.Path("bosh"), "delete-env", "--tty", testEnv.Path(manifest)}, extraDeployArgs...)
		_, _, exitCode, err := cmdRunner.RunStreamingCommand(multiWriter, cmdEnv, args...)

		if err != nil || exitCode != 0 {
			flushLog(quietCmdEnv["BOSH_LOG_PATH"])
		}

		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		return stdout.String()
	}

	var shutdownAgent = func() {
		_, _, exitCode, err := instanceSSH.RunCommandWithSudo("sv stop agent")
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
	}

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
			Fail("Failed to find stage start: " + stageName + ". Lines: start>>\n" + strings.Join(outputLines, "\n") + "<<end")
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

	BeforeSuite(func() {
		arch := "linux-amd64"
		if os.Getenv("BUILD_ARCH") != "" {
			arch = ""
			os.Setenv("GOOS", arch)
		}
		err := bitestutils.BuildExecutableForArch(arch)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, GinkgoWriter)
		fileSystem = boshsys.NewOsFileSystem(logger)

		var err error
		config, err = NewConfig(fileSystem)
		Expect(err).NotTo(HaveOccurred())

		err = config.Validate()
		Expect(err).NotTo(HaveOccurred())

		testEnv = NewTestEnvironment(
			fileSystem,
			logger,
		)

		extraDeployArgs = []string{"--vars-store", testEnv.Path("vars.yml")}

		cmdRunner = NewCmdRunner(logger)

		cmdEnv = map[string]string{
			"TMPDIR":         testEnv.Home(),
			"BOSH_LOG_LEVEL": "DEBUG",
			"BOSH_LOG_PATH":  testEnv.Path("bosh-init.log"),
		}
		quietCmdEnv = map[string]string{
			"TMPDIR":         testEnv.Home(),
			"BOSH_LOG_LEVEL": "ERROR",
			"BOSH_LOG_PATH":  testEnv.Path("bosh-init-cleanup.log"),
		}

		// clean up from previous failed tests
		deleteLogFile(cmdEnv["BOSH_LOG_PATH"])
		deleteLogFile(quietCmdEnv["BOSH_LOG_PATH"])

		boshCliPath := "./../out/bosh"
		Expect(fileSystem.FileExists(boshCliPath)).To(BeTrue())
		err = testEnv.Copy("bosh", boshCliPath)
		Expect(err).NotTo(HaveOccurred())

		instanceSSH = NewInstanceSSH(
			instanceUsername,
			instanceIP,
			instancePassword,
			fileSystem,
			logger,
		)

		err = testEnv.Copy("stemcell.tgz", config.StemcellPath)
		Expect(err).NotTo(HaveOccurred())

		err = testEnv.Copy("cpi-release.tgz", config.CPIReleasePath)
		Expect(err).NotTo(HaveOccurred())

		err = testEnv.Copy("sample-release-compiled.tgz", config.DummyCompiledReleasePath)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when deploying with a compiled release", func() {
		AfterEach(func() {
			flushLog(cmdEnv["BOSH_LOG_PATH"])
			deleteDeployment("test-manifest.yml", extraDeployArgs...)
		})

		It("is able to deploy given many variances with compiled releases", func() {
			updateCompiledReleaseDeploymentManifest("./assets/manifest.yml")

			By("deploying compiled releases successfully with expected output")
			stdout := deploy("test-manifest.yml", extraDeployArgs...)
			outputLines := strings.Split(stdout, "\n")
			numOutputLines := len(outputLines)

			doneIndex := 0
			stepIndex := -1
			nextStep := func() int { stepIndex++; return stepIndex }

			validatingSteps, doneIndex := findStage(outputLines, "validating", doneIndex)
			Expect(validatingSteps[nextStep()]).To(MatchRegexp("^  Validating release 'bosh-warden-cpi'" + stageFinishedPattern))
			Expect(validatingSteps[nextStep()]).To(MatchRegexp("^  Validating release 'sample-release'" + stageFinishedPattern))
			Expect(validatingSteps[nextStep()]).To(MatchRegexp("^  Validating cpi release" + stageFinishedPattern))
			Expect(validatingSteps[nextStep()]).To(MatchRegexp("^  Validating deployment manifest" + stageFinishedPattern))
			Expect(validatingSteps[nextStep()]).To(MatchRegexp("^  Validating stemcell" + stageFinishedPattern))

			installingSteps, doneIndex := findStage(outputLines, "installing CPI", doneIndex+1)
			numInstallingSteps := len(installingSteps)
			for _, line := range installingSteps[:numInstallingSteps-3] {
				Expect(line).To(MatchRegexp("^  Compiling package '.*/.*'" + stageFinishedPattern))
			}
			Expect(installingSteps[numInstallingSteps-3]).To(MatchRegexp("^  Installing packages" + stageFinishedPattern))
			Expect(installingSteps[numInstallingSteps-2]).To(MatchRegexp("^  Rendering job templates" + stageFinishedPattern))
			Expect(installingSteps[numInstallingSteps-1]).To(MatchRegexp("^  Installing job 'warden_cpi'" + stageFinishedPattern))

			deployingSteps, _ := findStage(outputLines, "deploying", doneIndex+1)
			numDeployingSteps := len(deployingSteps)
			Expect(deployingSteps[0]).To(MatchRegexp("^  Creating VM for instance 'dummy_instance_group/0' from stemcell '.*'" + stageFinishedPattern))
			Expect(deployingSteps[1]).To(MatchRegexp("^  Waiting for the agent on VM '.*' to be ready" + stageFinishedPattern))
			Expect(deployingSteps[2]).To(MatchRegexp("^  Creating disk" + stageFinishedPattern))
			Expect(deployingSteps[3]).To(MatchRegexp("^  Attaching disk '.*' to VM '.*'" + stageFinishedPattern))
			Expect(deployingSteps[4]).To(MatchRegexp("^  Rendering job templates" + stageFinishedPattern))

			for _, line := range deployingSteps[5 : numDeployingSteps-3] {
				Expect(line).To(MatchRegexp("^  Compiling package '.*/.*'" + stageCompiledPackageSkippedPattern))
			}

			Expect(deployingSteps[numDeployingSteps-3]).To(MatchRegexp("^  Updating instance 'dummy_instance_group/0'" + stageFinishedPattern))
			Expect(deployingSteps[numDeployingSteps-2]).To(MatchRegexp("^  Waiting for instance 'dummy_instance_group/0' to be running" + stageFinishedPattern))
			Expect(deployingSteps[numDeployingSteps-1]).To(MatchRegexp("^  Running the post-start scripts 'dummy_instance_group/0'" + stageFinishedPattern))

			Expect(outputLines[numOutputLines-4]).To(MatchRegexp("^Cleaning up rendered CPI jobs" + stageFinishedPattern))

			By("setting the ssh password")
			stdout, _, exitCode, err := instanceSSH.RunCommand("echo ssh-succeeded")
			Expect(err).ToNot(HaveOccurred())
			Expect(exitCode).To(Equal(0))
			Expect(stdout).To(ContainSubstring("ssh-succeeded"))

			By("skipping the deploy if there are no changes")
			stdout = deploy("test-manifest.yml", extraDeployArgs...)

			Expect(stdout).To(ContainSubstring("No deployment, stemcell or release changes. Skipping deploy."))
			Expect(stdout).ToNot(ContainSubstring("Started installing CPI jobs"))
			Expect(stdout).ToNot(ContainSubstring("Started deploying"))

			By("deleting the old VM if updating with a property change")
			updateCompiledReleaseDeploymentManifest("./assets/modified_manifest.yml")

			stdout = deploy("test-manifest.yml", extraDeployArgs...)

			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Stopping jobs on instance 'unknown/0'"))
			Expect(stdout).To(ContainSubstring("Unmounting disk"))

			Expect(stdout).ToNot(ContainSubstring("Creating disk"))

			By("deleting the agent when deploying without a working agent")
			shutdownAgent()
			updateCompiledReleaseDeploymentManifest("./assets/manifest.yml")

			stdout = deploy("test-manifest.yml", extraDeployArgs...)

			Expect(stdout).To(MatchRegexp("Waiting for the agent on VM '.*'\\.\\.\\. Failed " + stageTimePattern))
			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Creating VM for instance 'dummy_instance_group/0' from stemcell"))
			Expect(stdout).To(ContainSubstring("Finished deploying"))

			By("deleting all VMs, disks, and stemcells")
			stdout = deleteDeployment("test-manifest.yml", extraDeployArgs...)

			Expect(stdout).To(ContainSubstring("Stopping jobs on instance"))
			Expect(stdout).To(ContainSubstring("Deleting VM"))
			Expect(stdout).To(ContainSubstring("Deleting disk"))
			Expect(stdout).To(ContainSubstring("Deleting stemcell"))
			Expect(stdout).To(ContainSubstring("Finished deleting deployment"))
		})
	})

	Context("when deploying with a bogus mbus CA cert", func() {
		BeforeEach(func() {
			updateCompiledReleaseDeploymentManifest("./assets/manifest.yml")
		})

		AfterEach(func() {
			flushLog(cmdEnv["BOSH_LOG_PATH"])
			deleteDeployment("test-manifest.yml", extraDeployArgs...)
		})

		It("fails pinging the agent", func() {
			By("deploying with a bogus CA cert")
			fmt.Fprintf(GinkgoWriter, "\n--- DEPLOY ---\n")

			stdoutBuffer := &bytes.Buffer{}
			multiWriter := io.MultiWriter(stdoutBuffer, GinkgoWriter)

			_, _, exitCode, err := cmdRunner.RunStreamingCommand(
				multiWriter,
				cmdEnv,
				append(
					[]string{
						testEnv.Path("bosh"),
						"create-env", "--tty", testEnv.Path("test-manifest.yml"),
						"-o", "./assets/use-bogus-mbus-ca.yml",
					},
					extraDeployArgs...,
				)...,
			)
			Expect(err).To(HaveOccurred())
			Expect(exitCode).To(Equal(1))
			outputLines := strings.Split(stdoutBuffer.String(), "\n")

			Expect(outputLines).To(ContainElement(MatchRegexp("x509: certificate signed by unknown authority")))
		})
	})

	Context("when deploying with all network types", func() {
		AfterEach(func() {
			flushLog(cmdEnv["BOSH_LOG_PATH"])
			deleteDeployment("test-manifest.yml", extraDeployArgs...)
		})

		It("is successful", func() {
			updateCompiledReleaseDeploymentManifest("./assets/manifest_with_all_network_types.yml")

			stdout := deploy("test-manifest.yml", extraDeployArgs...)
			Expect(stdout).To(ContainSubstring("Finished deploying"))
		})
	})

	Context("When there is no deployment state to delete", func() {
		It("exits early", func() {
			updateCompiledReleaseDeploymentManifest("./assets/manifest.yml")
			stdout := deleteDeployment("test-manifest.yml", extraDeployArgs...)

			Expect(stdout).To(ContainSubstring("No deployment state file found"))
		})
	})
})
