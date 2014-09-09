package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var (
		command              bmcmd.Cmd
		config               bmconfig.Config
		fakeFs               *fakesys.FakeFileSystem
		fakeUI               *fakeui.FakeUI
		fakeExtractor        *testfakes.FakeMultiResponseExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator
		fakeReleaseCompiler  *fakebmcomp.FakeReleaseCompiler
		logger               boshlog.Logger
		deployment           bmdepl.Deployment
		release              bmrel.Release
		fakeRepo             *fakebmstemcell.FakeRepo
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		fakeReleaseCompiler = fakebmcomp.NewFakeReleaseCompiler()
		fakeRepo = fakebmstemcell.NewFakeRepo()

		logger = boshlog.NewLogger(boshlog.LevelNone)

		command = bmcmd.NewDeployCmd(
			fakeUI,
			config,
			fakeFs,
			fakeExtractor,
			fakeReleaseValidator,
			fakeReleaseCompiler,
			fakeRepo,
			logger,
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
							fakeRepo,
							logger,
						)

						release = bmrel.Release{
							Name:          "fake-release",
							Version:       "fake-version",
							ExtractedPath: "/some/release/path",
							TarballPath:   "/somepath",
						}

						releaseContents :=
							`---
name: fake-release
version: fake-version
`
						fakeFs.WriteFileString("/some/release/path/release.MF", releaseContents)
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
									fakeExtractor.SetDecompressBehavior("/somepath", "/some/release/path", nil)
									deployment = bmdepl.NewLocalDeployment("fake-deployment-name", map[string]interface{}{})
									fakeReleaseCompiler.SetCompileBehavior(release, "/some/deployment/file", nil)
									fakeRepo.SetSaveBehavior("/somestemcellpath", "/some/stemcell/path", bmstemcell.Stemcell{}, nil)
								})

								It("does not return an error", func() {
									err := runDeployCmd(command)
									Expect(err).NotTo(HaveOccurred())
								})

								It("compiles the release", func() {
									err := runDeployCmd(command)
									Expect(err).NotTo(HaveOccurred())
									Expect(fakeReleaseCompiler.CompileInputs[0].ManifestPath).To(Equal("/some/deployment/file"))
								})

								It("saves the stemcell", func() {
									fakeFs.WriteFile("/some/stemcell/path", []byte{})
									err := runDeployCmd(command)
									Expect(err).NotTo(HaveOccurred())
									Expect(fakeReleaseCompiler.CompileInputs[0].ManifestPath).To(Equal("/some/deployment/file"))
									Expect(fakeFs.FileExists("/some/stemcell/path")).To(BeFalse())
								})

								It("cleans up the extracted release directory", func() {
									err := runDeployCmd(command)
									Expect(err).NotTo(HaveOccurred())
									Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
								})
							})

							Context("and the tarball is not a valid BOSH release", func() {
								BeforeEach(func() {
									fakeExtractor.SetDecompressBehavior("/somepath", "/some/release/path", nil)
									fakeFs.WriteFileString("/some/release/path/release.MF", `{}`)
									fakeReleaseValidator.ValidateError = errors.New("fake-error")
								})

								It("returns err", func() {
									err := runDeployCmd(command)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("fake-error"))
								})

								It("cleans up the extracted release directory", func() {
									err := runDeployCmd(command)
									Expect(err).To(HaveOccurred())
									Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
								})
							})

							Context("and the tarball cannot be read", func() {
								It("returns err", func() {
									err := runDeployCmd(command)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("Reading CPI release from `/somepath'"))
									Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' is not a BOSH release"))
								})
							})

							Context("when compilation fails", func() {
								It("returns error", func() {
									fakeExtractor.SetDecompressBehavior("/somepath", "/some/release/path", nil)
									fakeReleaseCompiler.SetCompileBehavior(release, "/some/deployment/file", errors.New("fake-compile-error"))

									err := runDeployCmd(command)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("Compiling release"))
									Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
									Expect(fakeUI.Errors).To(ContainElement("Could not compile release"))
								})
							})

							Context("when reading stemcell fails", func() {
								It("returns error", func() {
									fakeExtractor.SetDecompressBehavior("/somepath", "/some/release/path", nil)
									fakeReleaseCompiler.SetCompileBehavior(release, "/some/deployment/file", nil)
									fakeRepo.SetSaveBehavior("/somestemcellpath", "/some/release/path", bmstemcell.Stemcell{}, errors.New("fake-reading-error"))

									err := runDeployCmd(command)
									Expect(err).To(HaveOccurred())
									Expect(err.Error()).To(ContainSubstring("Saving stemcell"))
									Expect(err.Error()).To(ContainSubstring("fake-reading-error"))
									Expect(fakeUI.Errors).To(ContainElement("Could not read stemcell"))
								})
							})
						})

						Context("when a extracted release path cannot be created", func() {
							BeforeEach(func() {
								fakeFs.TempDirError = errors.New("")
							})

							It("returns err", func() {
								err := runDeployCmd(command)
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Creating temp directory"))
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
								fakeRepo,
								logger,
							)
						})

						It("returns err", func() {
							err := runDeployCmd(command)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Reading deployment manifest for deploy"))
							Expect(fakeUI.Errors).To(ContainElement("Deployment manifest path `/some/deployment/file' does not exist"))
						})
					})
				})

				Context("when there is no deployment set", func() {
					It("returns err", func() {
						err := runDeployCmd(command)
						Expect(err).To(HaveOccurred())
						Expect(fakeUI.Errors).To(ContainElement("No deployment set"))
					})
				})
			})

			Context("When the CPI release file does not exist", func() {
				It("returns err when the CPI release file does not exist", func() {
					err := runDeployCmd(command)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Checking CPI release `/somepath' existence"))
					Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' does not exist"))
				})
			})
		})
	})
})

func runDeployCmd(command bmcmd.Cmd) error {
	return command.Run([]string{"/somepath", "/somestemcellpath"})
}
