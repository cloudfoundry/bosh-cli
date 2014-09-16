package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakedeploy "github.com/cloudfoundry/bosh-micro-cli/deployer/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var (
		command                    bmcmd.Cmd
		config                     bmconfig.Config
		fakeFs                     *fakesys.FakeFileSystem
		fakeUI                     *fakeui.FakeUI
		fakeCpiDeployer            *fakedeploy.FakeCpiDeployer
		logger                     boshlog.Logger
		release                    bmrel.Release
		fakeStemcellManager        *fakebmstemcell.FakeManager
		fakeStemcellManagerFactory *fakebmstemcell.FakeManagerFactory
		fakeCpiManifestParser      *fakebmdepl.FakeManifestParser

		stemcellTarballPath string
		expectedStemcellCID bmstemcell.CID
		expectedStemcell    bmstemcell.Stemcell
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}
		fakeCpiDeployer = fakedeploy.NewFakeCpiDeployer()
		fakeStemcellManager = fakebmstemcell.NewFakeManager()
		fakeStemcellManagerFactory = fakebmstemcell.NewFakeManagerFactory()
		fakeCpiManifestParser = fakebmdepl.NewFakeManifestParser()

		logger = boshlog.NewLogger(boshlog.LevelNone)
		command = bmcmd.NewDeployCmd(
			fakeUI,
			config,
			fakeFs,
			fakeCpiManifestParser,
			fakeCpiDeployer,
			fakeStemcellManagerFactory,
			logger,
		)

		stemcellTarballPath = "/stemcell/tarball/path"
		expectedStemcellCID = bmstemcell.CID("fake-stemcell-cid")
		expectedStemcell = bmstemcell.Stemcell{
			ImagePath:       "/stemcell/image/path",
			Name:            "fake-stemcell-name",
			Version:         "fake-stemcell-version",
			SHA1:            "fake-stemcell-sha1",
			CloudProperties: map[string]interface{}{},
		}
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

						// re-create command to update config.Deployment
						command = bmcmd.NewDeployCmd(
							fakeUI,
							config,
							fakeFs,
							fakeCpiManifestParser,
							fakeCpiDeployer,
							fakeStemcellManagerFactory,
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
							cloud      *fakebmcloud.FakeCloud
						)
						BeforeEach(func() {
							fakeFs.WriteFileString(config.Deployment, "")
							deployment = bmdepl.Deployment{}
							fakeCpiManifestParser.SetParseBehavior(config.Deployment, deployment, nil)
							cloud = fakebmcloud.NewFakeCloud()
							fakeCpiDeployer.SetDeployBehavior(deployment, release.TarballPath, cloud, nil)
							fakeStemcellManagerFactory.SetNewManagerBehavior(cloud, fakeStemcellManager)
							fakeStemcellManager.SetUploadBehavior(stemcellTarballPath, expectedStemcell, expectedStemcellCID, nil)
						})

						It("parses the CPI manifest", func() {
							err := command.Run([]string{release.TarballPath, stemcellTarballPath})
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeCpiManifestParser.ParseInputs[0].DeploymentPath).To(Equal("/some/deployment/file"))
						})

						It("deploys the CPI locally", func() {
							err := command.Run([]string{release.TarballPath, stemcellTarballPath})
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeCpiDeployer.DeployInputs[0].Deployment).To(Equal(deployment))
						})

						It("uploads the stemcell", func() {
							fakeFs.WriteFile(stemcellTarballPath, []byte{})
							err := command.Run([]string{release.TarballPath, stemcellTarballPath})
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeStemcellManager.UploadInputs).To(Equal(
								[]fakebmstemcell.UploadInput{
									fakebmstemcell.UploadInput{
										TarballPath: stemcellTarballPath,
									},
								},
							))
						})

						Context("when parsing the cpi deployment manifest fails", func() {
							It("returns error", func() {
								fakeCpiManifestParser.SetParseBehavior(config.Deployment, bmdepl.Deployment{}, errors.New("fake-parse-error"))

								err := command.Run([]string{release.TarballPath, stemcellTarballPath})
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Parsing CPI deployment manifest"))
								Expect(err.Error()).To(ContainSubstring("fake-parse-error"))
								Expect(fakeCpiManifestParser.ParseInputs[0].DeploymentPath).To(Equal("/some/deployment/file"))
							})
						})

						Context("when reading stemcell fails", func() {
							It("returns error", func() {
								fakeStemcellManager.SetUploadBehavior(stemcellTarballPath, bmstemcell.Stemcell{}, bmstemcell.CID(""), errors.New("fake-reading-error"))

								err := command.Run([]string{release.TarballPath, stemcellTarballPath})
								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("Uploading stemcell"))
								Expect(err.Error()).To(ContainSubstring("fake-reading-error"))
							})
						})
					})

					Context("when the deployment manifest is missing", func() {
						BeforeEach(func() {
							config.Deployment = "/some/deployment/file"

							// re-create command to update config.Deployment
							command = bmcmd.NewDeployCmd(
								fakeUI,
								config,
								fakeFs,
								fakeCpiManifestParser,
								fakeCpiDeployer,
								fakeStemcellManagerFactory,
								logger,
							)
						})

						It("returns err", func() {
							err := command.Run([]string{release.TarballPath, stemcellTarballPath})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Reading deployment manifest for deploy"))
							Expect(fakeUI.Errors).To(ContainElement("Deployment manifest path `/some/deployment/file' does not exist"))
						})
					})

				})

				Context("when there is no deployment set", func() {
					It("returns err", func() {
						err := command.Run([]string{release.TarballPath, stemcellTarballPath})
						Expect(err).To(HaveOccurred())
						Expect(fakeUI.Errors).To(ContainElement("No deployment set"))
					})
				})
			})

			Context("When the CPI release file does not exist", func() {
				It("returns err when the CPI release file does not exist", func() {
					err := command.Run([]string{release.TarballPath, stemcellTarballPath})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Checking CPI release `/somepath' existence"))
					Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' does not exist"))
				})
			})
		})
	})
})
