package cmd_test

import (
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command bmcmd.Cmd
		config  bmconfig.Config
		fakeFs  *fakesys.FakeFileSystem
		fakeUI  *fakeui.FakeUI
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}

		command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs)
	})

	Describe("Run", func() {
		Context("when there is a deployment set", func() {
			BeforeEach(func() {
				config.Deployment = "/some/deployment/file"
				command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs)
			})

			Context("when the deployment manifest exists", func() {
				BeforeEach(func() {
					fakeFs.WriteFileString(config.Deployment, "")
				})

				Context("when no arguments are given", func() {
					It("returns err", func() {
						err := command.Run([]string{})
						Expect(err).To(HaveOccurred())
						Expect(fakeUI.Errors).To(ContainElement("No CPI release provided"))
					})
				})

				Context("when a CPI release is given", func() {
					Context("and the CPI release is valid", func() {
						BeforeEach(func() {
							// TODO: make a valid CPI release
							fakeFs.WriteFileString("/somepath", "")
						})

						It("does not return an error", func() {
							err := command.Run([]string{"/somepath"})
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("and the CPI release does not exist", func() {
						It("returns err", func() {
							err := command.Run([]string{"/somepath"})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Validating CPI release"))
							Expect(fakeUI.Errors).To(ContainElement("CPI release '/somepath' does not exist"))
						})
					})
				})
			})

			Context("when the deployment manifest is missing", func() {
				BeforeEach(func() {
					config.Deployment = "/some/deployment/file"
					command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs)
				})

				It("returns err", func() {
					err := command.Run([]string{"/somepath"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading deployment manifest for deploy"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment manifest path '/some/deployment/file' does not exist"))
				})
			})
		})

		Context("when there is no deployment set", func() {
			It("returns err", func() {
				err := command.Run([]string{"/somepath"})
				Expect(err).To(HaveOccurred())
				Expect(fakeUI.Errors).To(ContainElement("No deployment set"))
			})
		})
	})
})
