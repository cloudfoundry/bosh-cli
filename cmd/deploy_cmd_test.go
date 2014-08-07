package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	faketar "github.com/cloudfoundry/bosh-micro-cli/tar/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command       bmcmd.Cmd
		config        bmconfig.Config
		fakeFs        *fakesys.FakeFileSystem
		fakeUI        *fakeui.FakeUI
		fakeExtractor *faketar.FakeExtractor
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}
		fakeExtractor = faketar.NewFakeExtractor()

		command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs, fakeExtractor)
	})

	Describe("Run", func() {
		Context("when there is a deployment set", func() {
			BeforeEach(func() {
				config.Deployment = "/some/deployment/file"
				command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs, fakeExtractor)
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
					Context("when a extracted release directory can be created", func() {
						BeforeEach(func() {
							fakeFs.TempDirDir = "/some/release/path"
						})

						Context("and the CPI release is valid", func() {
							BeforeEach(func() {
								fakeExtractor.AddExpectedArchive("/somepath")
								fakeFs.WriteFileString("/some/release/path/release.MF", `---
name: fake-release
version: fake-version
`)
							})

							It("does not return an error", func() {
								err := command.Run([]string{"/somepath"})
								Expect(err).NotTo(HaveOccurred())
							})

							It("cleans up the extracted release directory", func() {
								err := command.Run([]string{"/somepath"})
								Expect(err).NotTo(HaveOccurred())
								Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
							})
						})

						Context("and the CPI release is invalid", func() {
							BeforeEach(func() {
								fakeExtractor.AddExpectedArchive("/somepath")
								fakeFs.WriteFileString("/some/release/path/release.MF", `{}`)
							})

							It("returns err", func() {
								err := command.Run([]string{"/somepath"})
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Validating CPI release"))
								Expect(fakeUI.Errors).To(ContainElement("CPI release '/somepath' is not a valid BOSH release"))
							})

							It("cleans up the extracted release directory", func() {
								err := command.Run([]string{"/somepath"})
								Expect(err).To(HaveOccurred())
								Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
							})
						})

						Context("and the CPI release does not exist", func() {
							It("returns err", func() {
								err := command.Run([]string{"/somepath"})
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Reading CPI release from '/somepath'"))
								Expect(fakeUI.Errors).To(ContainElement("CPI release '/somepath' is not a BOSH release"))
							})
						})
					})

					Context("when a extracted release path cannot be created", func() {
						BeforeEach(func() {
							fakeFs.TempDirError = errors.New("")
						})

						It("returns err", func() {
							err := command.Run([]string{"/somepath"})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Creating extracted release path"))
							Expect(fakeUI.Errors).To(ContainElement("Could not create a temporary directory"))
						})
					})
				})
			})

			Context("when the deployment manifest is missing", func() {
				BeforeEach(func() {
					config.Deployment = "/some/deployment/file"
					command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs, fakeExtractor)
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
