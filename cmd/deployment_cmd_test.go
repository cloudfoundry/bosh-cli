package cmd_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command           Cmd
		deploymentFile    bmconfig.DeploymentFile
		userConfig        bmconfig.UserConfig
		userConfigService bmconfig.UserConfigService
		manifestPath      string
		fakeUI            *fakebmui.FakeUI
		fakeFs            *fakesys.FakeFileSystem
		fakeUUID          *fakeuuid.FakeGenerator
		logger            boshlog.Logger
	)

	BeforeEach(func() {
		fakeUI = &fakebmui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		userConfigService = bmconfig.NewFileSystemUserConfigService("/fake-user-config", fakeFs, logger)
		fakeUUID = &fakeuuid.FakeGenerator{}
		deploymentFile = bmconfig.DeploymentFile{}

		command = NewDeploymentCmd(
			fakeUI,
			userConfig,
			userConfigService,
			deploymentFile,
			fakeFs,
			fakeUUID,
			logger,
		)
	})

	Context("Run", func() {
		Context("ran with valid args", func() {
			Context("when the deployment manifest exists", func() {
				BeforeEach(func() {
					fakeUUID.GeneratedUuid = "abc123"
					manifestDir, err := fakeFs.TempDir("deployment-cmd")
					Expect(err).ToNot(HaveOccurred())

					manifestPath = path.Join("/", manifestDir, "manifestFile.yml")
					err = fakeFs.WriteFileString(manifestPath, "")
					Expect(err).ToNot(HaveOccurred())
				})

				It("says 'deployment set..' to the UI", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement(ContainSubstring(fmt.Sprintf("Deployment set to `%s'", manifestPath))))
				})

				It("saves the deployment manifest to the user config", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())

					userConfigContents, err := fakeFs.ReadFile("/fake-user-config")
					Expect(err).NotTo(HaveOccurred())
					userConfig := bmconfig.UserConfig{}
					err = json.Unmarshal(userConfigContents, &userConfig)
					Expect(err).NotTo(HaveOccurred())

					Expect(userConfig).To(Equal(bmconfig.UserConfig{DeploymentFile: manifestPath}))
				})

				It("saves absolute path to deployment manifest in user config", func() {
					wd, err := os.Getwd()
					Expect(err).NotTo(HaveOccurred())
					manifestAbsolutePath := path.Join(wd, "fake-manifest-file")

					err = fakeFs.WriteFileString(manifestAbsolutePath, "")
					Expect(err).NotTo(HaveOccurred())

					err = command.Run([]string{"fake-manifest-file"})
					Expect(err).NotTo(HaveOccurred())

					userConfigContents, err := fakeFs.ReadFile("/fake-user-config")
					Expect(err).NotTo(HaveOccurred())
					userConfig := bmconfig.UserConfig{}
					err = json.Unmarshal(userConfigContents, &userConfig)
					Expect(err).NotTo(HaveOccurred())

					Expect(userConfig).To(Equal(bmconfig.UserConfig{DeploymentFile: manifestAbsolutePath}))
				})

				It("creates a deployment config", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())

					userConfig := bmconfig.UserConfig{DeploymentFile: manifestPath}
					deploymentConfigService := bmconfig.NewFileSystemDeploymentConfigService(userConfig.DeploymentConfigFilePath(), fakeFs, logger)
					deploymentConfig, err := deploymentConfigService.Load()
					Expect(err).NotTo(HaveOccurred())

					Expect(deploymentConfig).To(Equal(bmconfig.DeploymentFile{UUID: "abc123"}))
				})

				It("reuses the existing deployment config if it exists", func() {
					userConfig := bmconfig.UserConfig{DeploymentFile: manifestPath}
					deploymentConfigService := bmconfig.NewFileSystemDeploymentConfigService(
						userConfig.DeploymentConfigFilePath(),
						fakeFs,
						logger,
					)
					deploymentConfigService.Save(bmconfig.DeploymentFile{UUID: "def456"})

					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())

					deploymentConfig, err := deploymentConfigService.Load()
					Expect(err).NotTo(HaveOccurred())

					Expect(deploymentConfig).To(Equal(bmconfig.DeploymentFile{UUID: "def456"}))
				})
			})

			Context("when the deployment manifest does not exist", func() {
				It("returns err", func() {
					err := command.Run([]string{"/fake/manifest/path"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Verifying that the deployment `/fake/manifest/path' exists"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment `/fake/manifest/path' does not exist"))
				})
			})
		})

		Context("ran without args", func() {
			Context("a deployment manifest is present in the config", func() {
				BeforeEach(func() {
					userConfig := bmconfig.UserConfig{DeploymentFile: "/somepath"}
					command = NewDeploymentCmd(fakeUI,
						userConfig,
						userConfigService,
						deploymentFile,
						fakeFs,
						fakeUUID,
						logger,
					)
				})

				It("says `Deployment set to '<manifest_path>'`", func() {
					err := command.Run([]string{})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement("Current deployment is `/somepath'"))
				})
			})

			Context("no deployment manifest is present in the config", func() {
				It("says `No deployment set`", func() {
					err := command.Run([]string{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("No deployment set"))
					Expect(fakeUI.Errors).To(ContainElement("No deployment set"))
				})
			})
		})
	})
})
