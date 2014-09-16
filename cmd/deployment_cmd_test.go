package cmd_test

import (
	"fmt"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command      Cmd
		fakeService  *fakebmconfig.FakeService
		manifestPath string
		fakeUI       *fakebmui.FakeUI
		fakeFs       *fakesys.FakeFileSystem
		fakeUUID     *fakeuuid.FakeGenerator
		logger       boshlog.Logger
		config       bmconfig.Config
	)

	BeforeEach(func() {
		fakeUI = &fakebmui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		fakeService = fakebmconfig.NewFakeService()
		fakeUUID = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		config = bmconfig.Config{
			ContainingDir: "/fake-path",
		}

		command = NewDeploymentCmd(
			fakeUI,
			config,
			fakeService,
			fakeFs,
			fakeUUID,
			logger,
		)
	})

	Context("#Run", func() {
		Context("ran with valid args", func() {
			Context("when the deployment manifest exists", func() {
				var (
					expectedConfig bmconfig.Config
				)
				BeforeEach(func() {
					fakeUUID.GeneratedUuid = "abc123"
					manifestDir, err := fakeFs.TempDir("deployment-cmd")
					Expect(err).ToNot(HaveOccurred())

					manifestPath = path.Join(manifestDir, "manifestFile.yml")
					err = fakeFs.WriteFileString(manifestPath, "")
					Expect(err).ToNot(HaveOccurred())

					expectedConfig = bmconfig.Config{
						Deployment:     manifestPath,
						DeploymentUUID: "abc123",
						ContainingDir:  "/fake-path",
					}
					fakeService.SetSaveBehavior(expectedConfig, nil)
				})

				It("says 'deployment set..' to the UI", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement(ContainSubstring(fmt.Sprintf("Deployment set to `%s'", manifestPath))))
				})

				It("saves the deployment manifest in the config", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeService.SaveInputs).To(Equal(
						[]fakebmconfig.SaveInput{
							fakebmconfig.SaveInput{
								Config: expectedConfig,
							},
						},
					))
				})

				It("creates the blobstore folder", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeFs.FileExists("/fake-path/.bosh_micro/abc123/blobs")).To(BeTrue())
				})
			})

			Context("when the deployment manifest does not exist", func() {
				It("returns err", func() {
					err := command.Run([]string{"fake/manifest/path"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Setting deployment manifest"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment `fake/manifest/path' does not exist"))
				})
			})
		})

		Context("ran without args", func() {
			Context("a deployment manifest is present in the config", func() {
				BeforeEach(func() {
					config := bmconfig.Config{Deployment: "/somepath"}
					command = NewDeploymentCmd(fakeUI, config, fakeService, fakeFs, fakeUUID, logger)
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
