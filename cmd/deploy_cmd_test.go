package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	faketar "github.com/cloudfoundry/bosh-micro-cli/tar/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command              bmcmd.Cmd
		config               bmconfig.Config
		fakeFs               *fakesys.FakeFileSystem
		fakeUI               *fakeui.FakeUI
		fakeExtractor        *faketar.FakeExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator
		fakeReleaseCompiler  *fakebmcomp.FakeReleaseCompiler
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}
		fakeExtractor = faketar.NewFakeExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		fakeReleaseCompiler = fakebmcomp.NewFakeReleaseCompiler()

		command = bmcmd.NewDeployCmd(
			fakeUI,
			config,
			fakeFs,
			fakeExtractor,
			fakeReleaseValidator,
			fakeReleaseCompiler,
		)
	})

	Describe("Run", func() {
		Context("when no arguments are given", func() {
			It("returns err", func() {
				err := command.Run([]string{})
				Expect(err).To(HaveOccurred())
				Expect(fakeUI.Errors).To(ContainElement("No CPI release provided"))
			})
		})

		Context("when a CPI release is given", func() {
			Context("When the CPI release file exists", func() {
				BeforeEach(func() {
					fakeFs.WriteFileString("/somepath", "")
				})

				Context("when there is a deployment set", func() {
					BeforeEach(func() {
						config.Deployment = "/some/deployment/file"
						command = bmcmd.NewDeployCmd(
							fakeUI,
							config,
							fakeFs,
							fakeExtractor,
							fakeReleaseValidator,
							fakeReleaseCompiler,
						)
					})

					Context("when the deployment manifest exists", func() {
						BeforeEach(func() {
							fakeFs.WriteFileString(config.Deployment, "")
						})

						Context("when a extracted release directory can be created", func() {
							BeforeEach(func() {
								fakeFs.TempDirDir = "/some/release/path"
							})

							Context("and the tarball is a valid BOSH release", func() {
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

								It("compiles the release", func() {
									err := command.Run([]string{"/somepath"})
									Expect(err).NotTo(HaveOccurred())
									Expect(fakeReleaseCompiler.CompileRelease.Name).To(Equal("fake-release"))
								})

								It("cleans up the extracted release directory", func() {
									err := command.Run([]string{"/somepath"})
									Expect(err).NotTo(HaveOccurred())
									Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
								})
							})

							Context("and the tarball is not a valid BOSH release", func() {
								BeforeEach(func() {
									fakeExtractor.AddExpectedArchive("/somepath")
									fakeFs.WriteFileString("/some/release/path/release.MF", `{}`)
									fakeReleaseValidator.ValidateError = errors.New("fake-error")
								})

								It("returns err", func() {
									err := command.Run([]string{"/somepath"})
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("fake-error"))
								})

								It("cleans up the extracted release directory", func() {
									err := command.Run([]string{"/somepath"})
									Expect(err).To(HaveOccurred())
									Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
								})
							})

							Context("and the tarball cannot be read", func() {
								It("returns err", func() {
									err := command.Run([]string{"/somepath"})
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("Reading CPI release from `/somepath'"))
									Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' is not a BOSH release"))
								})
							})

							Context("when compilation fails", func() {
								It("returns error", func() {
									fakeExtractor.AddExpectedArchive("/somepath")
									fakeFs.WriteFileString("/some/release/path/release.MF", `{}`)
									fakeReleaseCompiler.CompileError = errors.New("fake-error-compile")
									err := command.Run([]string{"/somepath"})
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("Compiling release"))
									Expect(fakeUI.Errors).To(ContainElement("Could not compile release"))
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

					Context("when the deployment manifest is missing", func() {
						BeforeEach(func() {
							config.Deployment = "/some/deployment/file"
							command = bmcmd.NewDeployCmd(
								fakeUI,
								config,
								fakeFs,
								fakeExtractor,
								fakeReleaseValidator,
								fakeReleaseCompiler,
							)
						})

						It("returns err", func() {
							err := command.Run([]string{"/somepath"})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Reading deployment manifest for deploy"))
							Expect(fakeUI.Errors).To(ContainElement("Deployment manifest path `/some/deployment/file' does not exist"))
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

			Context("When the CPI release file does not exist", func() {
				It("returns err when the CPI release file does not exist", func() {
					err := command.Run([]string{"/somepath"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Checking CPI release `/somepath' existence"))
					Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' does not exist"))
				})
			})
		})
	})
})
