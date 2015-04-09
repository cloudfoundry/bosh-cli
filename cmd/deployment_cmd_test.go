package cmd_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	biconfig "github.com/cloudfoundry/bosh-init/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"

	. "github.com/cloudfoundry/bosh-init/cmd"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command           Cmd
		userConfig        biconfig.UserConfig
		userConfigService biconfig.UserConfigService
		manifestPath      string
		fakeUI            *fakebiui.FakeUI
		fakeFs            *fakesys.FakeFileSystem
		fakeUUID          *fakeuuid.FakeGenerator
		logger            boshlog.Logger
		fakeStage         *fakebiui.FakeStage
	)

	BeforeEach(func() {
		fakeUI = &fakebiui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		userConfigService = biconfig.NewFileSystemUserConfigService("/fake-user-config", fakeFs, logger)
		fakeUUID = &fakeuuid.FakeGenerator{}

		fakeStage = fakebiui.NewFakeStage()

		command = NewDeploymentCmd(
			fakeUI,
			userConfig,
			userConfigService,
			fakeFs,
			fakeUUID,
			logger,
		)
	})

	Context("Run", func() {
		Context("ran with valid args", func() {
			Context("when the deployment manifest exists", func() {
				BeforeEach(func() {
					manifestDir, err := fakeFs.TempDir("deployment-cmd")
					Expect(err).ToNot(HaveOccurred())

					manifestPath = path.Join("/", manifestDir, "manifestFile.yml")
					err = fakeFs.WriteFileString(manifestPath, "")
					Expect(err).ToNot(HaveOccurred())
				})

				It("prints confirmation with full path to the UI", func() {
					err := command.Run(fakeStage, []string{manifestPath})
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement(ContainSubstring(fmt.Sprintf("Deployment manifest set to '%s'", manifestPath))))
				})

				It("saves the deployment manifest to the user config", func() {
					err := command.Run(fakeStage, []string{manifestPath})
					Expect(err).ToNot(HaveOccurred())

					userConfigContents, err := fakeFs.ReadFile("/fake-user-config")
					Expect(err).ToNot(HaveOccurred())
					userConfig := biconfig.UserConfig{}
					err = json.Unmarshal(userConfigContents, &userConfig)
					Expect(err).ToNot(HaveOccurred())

					Expect(userConfig).To(Equal(biconfig.UserConfig{DeploymentManifestPath: manifestPath}))
				})

				It("saves absolute path to deployment manifest in user config", func() {
					wd, err := os.Getwd()
					Expect(err).ToNot(HaveOccurred())
					manifestAbsolutePath := path.Join(wd, "fake-manifest-file")

					err = fakeFs.WriteFileString(manifestAbsolutePath, "")
					Expect(err).ToNot(HaveOccurred())

					err = command.Run(fakeStage, []string{"fake-manifest-file"})
					Expect(err).ToNot(HaveOccurred())

					userConfigContents, err := fakeFs.ReadFile("/fake-user-config")
					Expect(err).ToNot(HaveOccurred())
					userConfig := biconfig.UserConfig{}
					err = json.Unmarshal(userConfigContents, &userConfig)
					Expect(err).ToNot(HaveOccurred())

					Expect(userConfig).To(Equal(biconfig.UserConfig{DeploymentManifestPath: manifestAbsolutePath}))
				})

				It("creates a deployment config", func() {
					err := command.Run(fakeStage, []string{manifestPath})
					Expect(err).ToNot(HaveOccurred())

					userConfig := biconfig.UserConfig{DeploymentManifestPath: manifestPath}
					deploymentConfigService := biconfig.NewFileSystemDeploymentConfigService(userConfig.DeploymentConfigPath(), fakeFs, fakeUUID, logger)
					deploymentConfig, err := deploymentConfigService.Load()
					Expect(err).ToNot(HaveOccurred())

					Expect(deploymentConfig).To(Equal(biconfig.DeploymentFile{
						DirectorID: "fake-uuid-0",
					}))
				})

				It("reuses the existing deployment config if it exists", func() {
					userConfig := biconfig.UserConfig{DeploymentManifestPath: manifestPath}
					deploymentConfigService := biconfig.NewFileSystemDeploymentConfigService(
						userConfig.DeploymentConfigPath(),
						fakeFs,
						fakeUUID,
						logger,
					)
					deploymentConfigService.Save(biconfig.DeploymentFile{
						DirectorID:     "fake-director-id",
						InstallationID: "fake-installation-id",
					})

					err := command.Run(fakeStage, []string{manifestPath})
					Expect(err).ToNot(HaveOccurred())

					deploymentConfig, err := deploymentConfigService.Load()
					Expect(err).ToNot(HaveOccurred())

					Expect(deploymentConfig).To(Equal(biconfig.DeploymentFile{
						DirectorID:     "fake-director-id",
						InstallationID: "fake-installation-id",
					}))
				})
			})

			Context("when the deployment manifest does not exist", func() {
				It("returns err", func() {
					err := command.Run(fakeStage, []string{"/fake/manifest/path"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Verifying that the deployment '/fake/manifest/path' exists"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment '/fake/manifest/path' does not exist"))
				})
			})
		})

		Context("ran without args", func() {
			Context("a deployment manifest is present in the config", func() {
				BeforeEach(func() {
					userConfig := biconfig.UserConfig{DeploymentManifestPath: "/path/to/manifest.yml"}
					command = NewDeploymentCmd(fakeUI,
						userConfig,
						userConfigService,
						fakeFs,
						fakeUUID,
						logger,
					)
				})

				Context("when the manifest file exists", func() {
					BeforeEach(func() {
						err := fakeFs.WriteFileString("/path/to/manifest.yml", "fake-manifest-contents")
						Expect(err).ToNot(HaveOccurred())
					})

					It("prints the manifest path to the ui", func() {
						err := command.Run(fakeStage, []string{})
						Expect(err).ToNot(HaveOccurred())
						Expect(fakeUI.Said).To(ContainElement("Deployment manifest: '/path/to/manifest.yml'"))
						Expect(fakeUI.Said).To(ContainElement("Deployment state: '/path/to/deployment.json'"))
					})
				})

				Context("when the manifest file does not exist", func() {
					It("prints to the ui & returns an error", func() {
						err := command.Run(fakeStage, []string{})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Running deployment cmd: Deployment manifest does not exist at '/path/to/manifest.yml'"))
						Expect(fakeUI.Errors).To(ContainElement("Deployment manifest does not exist"))
					})
				})
			})

			Context("when no deployment manifest is present in the config", func() {
				It("prints to the ui & returns an error", func() {
					err := command.Run(fakeStage, []string{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Running deployment cmd: Deployment manifest not set"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment manifest not set"))
				})
			})
		})
	})
})
