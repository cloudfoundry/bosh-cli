package cmd_test

import (
	"fmt"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	fakeconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command      Cmd
		fakeService  *fakeconfig.FakeService
		manifestPath string
		fakeUI       *fakeui.FakeUI
		fakeFs       *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		fakeService = &fakeconfig.FakeService{}

		command = NewDeploymentCmd(fakeUI, bmconfig.Config{}, fakeService, fakeFs)
	})

	Context("#Run", func() {
		Context("ran with valid args", func() {
			Context("when the deployment manifest exists", func() {
				BeforeEach(func() {
					file, err := fakeFs.TempFile("bosh-micro-cli-manifest")
					Expect(err).ToNot(HaveOccurred())
					manifestPath = file.Name()
				})

				It("says 'deployment set..' to the UI", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeUI.Said).To(ContainElement(ContainSubstring(fmt.Sprintf("Deployment set to `%s'", manifestPath))))
				})

				It("saves the deployment manifest in the config", func() {
					err := command.Run([]string{manifestPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeService.Saved).To(Equal(bmconfig.Config{Deployment: manifestPath}))
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
					command = NewDeploymentCmd(fakeUI, config, fakeService, fakeFs)
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
