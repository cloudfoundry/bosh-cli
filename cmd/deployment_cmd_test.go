package cmd_test

import (
	"fmt"
	"path"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	fakeconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
	fakews "github.com/cloudfoundry/bosh-micro-cli/workspace/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command      Cmd
		fakeService  *fakeconfig.FakeService
		manifestPath string
		fakeUI       *fakeui.FakeUI
		fakeFs       *fakesys.FakeFileSystem
		fakeWs       *fakews.FakeWorkspace
		fakeUUID     *fakeuuid.FakeGenerator
		logger       boshlog.Logger
		config       bmconfig.Config
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		fakeService = &fakeconfig.FakeService{}
		fakeUUID = &fakeuuid.FakeGenerator{}
		fakeWs = fakews.NewFakeWorkspace()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		config = bmconfig.Config{}

		command = NewDeploymentCmd(
			fakeUI,
			config,
			fakeService,
			fakeFs,
			fakeWs,
			fakeUUID,
			logger,
		)
	})

	Context("#Run", func() {
		Context("ran with valid args", func() {
			Context("when the deployment manifest exists", func() {
				BeforeEach(func() {
					fakeUUID.GeneratedUuid = "abc123"
					manifestDir, err := fakeFs.TempDir("deployment-cmd")
					Expect(err).ToNot(HaveOccurred())

					manifestPath = path.Join(manifestDir, "manifestFile.yml")
					err = fakeFs.WriteFileString(manifestPath, "")
					Expect(err).ToNot(HaveOccurred())
				})

				It("says 'deployment set..' to the UI", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement(ContainSubstring(fmt.Sprintf("Deployment set to `%s'", manifestPath))))
				})

				It("saves the deployment manifest in the config", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeService.Saved).To(Equal(bmconfig.Config{
						Deployment:     manifestPath,
						DeploymentUUID: "abc123",
					}))
				})

				It("initializes the workspace with the correct uuid", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeWs.InitializeCalled).To(BeTrue())
					Expect(fakeWs.InitializeUUID).To(Equal("abc123"))
				})
			})

			Context("when the deployment manifest does not exist", func() {
				It("returns err", func() {
					err := command.Run([]string{"fake/manifest/path"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Setting deployment manifest"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment `fake/manifest/path' does not exist"))
				})

				It("does not initializes the workspace", func() {
					err := command.Run([]string{"fake/manifest/path"})
					Expect(err).To(HaveOccurred())

					Expect(fakeWs.InitializeCalled).To(BeFalse())
				})
			})
		})

		Context("ran without args", func() {
			Context("a deployment manifest is present in the config", func() {
				BeforeEach(func() {
					config := bmconfig.Config{Deployment: "/somepath"}
					command = NewDeploymentCmd(fakeUI, config, fakeService, fakeFs, fakeWs, fakeUUID, logger)
				})

				It("says `Deployment set to '<manifest_path>'`", func() {
					err := command.Run([]string{})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement("Current deployment is `/somepath'"))
				})

				It("does not initializes the workspace", func() {
					err := command.Run([]string{})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeWs.InitializeCalled).To(BeFalse())
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
