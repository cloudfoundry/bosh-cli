package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeploy "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakedeploy "github.com/cloudfoundry/bosh-micro-cli/deployer/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var (
		command               bmcmd.Cmd
		config                bmconfig.Config
		fakeFs                *fakesys.FakeFileSystem
		fakeUI                *fakeui.FakeUI
		fakeCpiDeployer       *fakedeploy.FakeCpiDeployer
		logger                boshlog.Logger
		release               bmrel.Release
		fakeRepo              *fakebmstemcell.FakeRepo
		fakeCpiManifestParser *fakebmdepl.FakeManifestParser
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}
		fakeCpiDeployer = fakedeploy.NewFakeCpiDeployer()
		fakeRepo = fakebmstemcell.NewFakeRepo()
		fakeCpiManifestParser = fakebmdepl.NewFakeManifestParser()

		logger = boshlog.NewLogger(boshlog.LevelNone)
		command = bmcmd.NewDeployCmd(
			fakeUI,
			config,
			fakeFs,
			fakeCpiManifestParser,
			fakeCpiDeployer,
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
							fakeCpiManifestParser,
							fakeCpiDeployer,
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
						var (
							deployment bmdepl.Deployment
						)
						BeforeEach(func() {
							fakeFs.WriteFileString(config.Deployment, "")
							deployment = bmdepl.Deployment{}
							fakeCpiManifestParser.SetParseBehavior(config.Deployment, deployment, nil)
							fakeCpiDeployer.SetDeployBehavior(deployment, "/somepath", bmdeploy.Cloud{}, nil)
							fakeRepo.SetSaveBehavior("/somestemcellpath", "/some/stemcell/path", bmstemcell.Stemcell{}, nil)
						})

						It("parses the CPI manifest", func() {
							err := runDeployCmd(command)
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeCpiManifestParser.ParseInputs[0].DeploymentPath).To(Equal("/some/deployment/file"))
						})

						It("deploys the CPI locally", func() {
							err := runDeployCmd(command)
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeCpiDeployer.DeployInputs[0].Deployment).To(Equal(deployment))
						})

						It("saves the stemcell and cleans up the temp path", func() {
							fakeFs.WriteFile("/some/stemcell/path", []byte{})
							err := runDeployCmd(command)
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeFs.FileExists("/some/stemcell/path")).To(BeFalse())
						})

						Context("when parsing the cpi deployment manifest fails", func() {
							It("returns error", func() {
								fakeCpiManifestParser.SetParseBehavior(config.Deployment, bmdepl.Deployment{}, errors.New("fake-parse-error"))

								err := runDeployCmd(command)
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Parsing CPI deployment manifest"))
								Expect(err.Error()).To(ContainSubstring("fake-parse-error"))
								Expect(fakeCpiManifestParser.ParseInputs[0].DeploymentPath).To(Equal("/some/deployment/file"))
							})
						})

						Context("when reading stemcell fails", func() {
							It("returns error", func() {
								fakeRepo.SetSaveBehavior("/somestemcellpath", "", bmstemcell.Stemcell{}, errors.New("fake-reading-error"))

								err := runDeployCmd(command)
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Saving stemcell"))
								Expect(err.Error()).To(ContainSubstring("fake-reading-error"))
								Expect(fakeUI.Errors).To(ContainElement("Could not read stemcell"))
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
								fakeCpiManifestParser,
								fakeCpiDeployer,
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
