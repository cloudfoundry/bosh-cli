package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi/fakes"
	fakebmdeployer "github.com/cloudfoundry/bosh-micro-cli/deployer/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var (
		command               bmcmd.Cmd
		userConfig            bmconfig.UserConfig
		fakeFs                *fakesys.FakeFileSystem
		fakeUI                *fakeui.FakeUI
		fakeCPIInstaller      *fakebmcpi.FakeInstaller
		fakeCPIRelease        *fakebmrel.FakeRelease
		logger                boshlog.Logger
		release               bmrel.Release
		fakeStemcellExtractor *fakebmstemcell.FakeExtractor

		fakeDeployer         *fakebmdeployer.FakeDeployer
		fakeDeploymentRecord *fakebmdeployer.FakeDeploymentRecord

		fakeCpiDeploymentParser *fakebmdepl.FakeCPIDeploymentParser
		fakeBoshManifestParser  *fakebmdepl.FakeManifestParser
		fakeDeploymentValidator *fakebmdeplval.FakeValidator

		fakeCompressor    *fakecmd.FakeCompressor
		fakeJobRenderer   *fakebmtemp.FakeJobRenderer
		fakeUUIDGenerator *fakeuuid.FakeGenerator

		fakeEventLogger *fakebmlog.FakeEventLogger
		fakeStage       *fakebmlog.FakeStage

		deploymentManifestPath    string
		cpiReleaseTarballPath     string
		stemcellTarballPath       string
		expectedExtractedStemcell bmstemcell.ExtractedStemcell
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		deploymentManifestPath = "/some/deployment/file"
		userConfig = bmconfig.UserConfig{
			DeploymentFile: deploymentManifestPath,
		}
		fakeFs.WriteFileString(deploymentManifestPath, "")

		fakeCPIInstaller = fakebmcpi.NewFakeInstaller()
		fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()

		fakeDeployer = fakebmdeployer.NewFakeDeployer()

		fakeCpiDeploymentParser = fakebmdepl.NewFakeCPIDeploymentParser()
		fakeBoshManifestParser = fakebmdepl.NewFakeManifestParser()
		fakeDeploymentValidator = fakebmdeplval.NewFakeValidator()

		fakeEventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior(fakeStage)

		fakeCompressor = fakecmd.NewFakeCompressor()
		fakeJobRenderer = fakebmtemp.NewFakeJobRenderer()
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}

		fakeDeploymentRecord = fakebmdeployer.NewFakeDeploymentRecord()

		logger = boshlog.NewLogger(boshlog.LevelNone)
		command = bmcmd.NewDeployCmd(
			fakeUI,
			userConfig,
			fakeFs,
			fakeCpiDeploymentParser,
			fakeBoshManifestParser,
			fakeDeploymentValidator,
			fakeCPIInstaller,
			fakeStemcellExtractor,
			fakeDeploymentRecord,
			fakeDeployer,
			fakeEventLogger,
			logger,
		)

		cpiReleaseTarballPath = "/release/tarball/path"

		stemcellTarballPath = "/stemcell/tarball/path"
		expectedExtractedStemcell = bmstemcell.NewExtractedStemcell(
			bmstemcell.Manifest{
				ImagePath:          "/stemcell/image/path",
				Name:               "fake-stemcell-name",
				Version:            "fake-stemcell-version",
				SHA1:               "fake-stemcell-sha1",
				RawCloudProperties: map[interface{}]interface{}{},
			},
			bmstemcell.ApplySpec{},
			"fake-extracted-path",
			fakeFs,
		)
	})

	Describe("Run", func() {
		It("returns err when no arguments are given", func() {
			err := command.Run([]string{})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires exactly 2 arguments"))
		})

		It("returns err when 1 argument is given", func() {
			err := command.Run([]string{"something"})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires exactly 2 arguments"))
		})

		It("returns err when 3 arguments are given", func() {
			err := command.Run([]string{"a", "b", "c"})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires exactly 2 arguments"))
		})

		Context("when a CPI release is given", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(cpiReleaseTarballPath, "")
				fakeFs.WriteFileString(stemcellTarballPath, "")
			})

			Context("when there is a deployment set", func() {
				BeforeEach(func() {
					userConfig.DeploymentFile = deploymentManifestPath

					// re-create command to update userConfig.DeploymentFile
					command = bmcmd.NewDeployCmd(
						fakeUI,
						userConfig,
						fakeFs,
						fakeCpiDeploymentParser,
						fakeBoshManifestParser,
						fakeDeploymentValidator,
						fakeCPIInstaller,
						fakeStemcellExtractor,
						fakeDeploymentRecord,
						fakeDeployer,
						fakeEventLogger,
						logger,
					)

					release = bmrel.NewRelease(
						"fake-release",
						"fake-version",
						[]bmrel.Job{},
						[]*bmrel.Package{},
						"/some/release/path",
						fakeFs,
					)

					releaseContents :=
						`---
name: fake-release
version: fake-version
`
					fakeFs.WriteFileString("/some/release/path/release.MF", releaseContents)
					fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
						{
							Err: nil,
						},
					})
				})

				Context("when the deployment manifest exists", func() {
					var (
						boshDeployment bmdepl.Deployment
						cpiDeployment  bmdepl.CPIDeployment
						cloud          *fakebmcloud.FakeCloud
					)
					BeforeEach(func() {
						fakeFs.WriteFileString(userConfig.DeploymentFile, "")
						cpiDeployment = bmdepl.CPIDeployment{
							Registry: bmdepl.Registry{
								Username: "fake-username",
							},
							SSHTunnel: bmdepl.SSHTunnel{
								Host: "fake-host",
							},
							Mbus: "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint",
						}
						fakeCpiDeploymentParser.ParseDeployment = cpiDeployment

						boshDeployment = bmdepl.Deployment{
							Name: "fake-deployment-name",
							Jobs: []bmdepl.Job{
								{
									Name: "fake-job-name",
								},
							},
						}
						fakeBoshManifestParser.SetParseBehavior(userConfig.DeploymentFile, boshDeployment, nil)
						cloud = fakebmcloud.NewFakeCloud()
						fakeCPIRelease = fakebmrel.NewFakeRelease()
						fakeCPIInstaller.SetExtractBehavior(cpiReleaseTarballPath, fakeCPIRelease, nil)
						fakeCPIInstaller.SetInstallBehavior(cpiDeployment, fakeCPIRelease, cloud, nil)

						fakeDeployer.SetDeployBehavior(nil)
						fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, expectedExtractedStemcell, nil)

						fakeFs.WriteFile(stemcellTarballPath, []byte{})

						fakeDeploymentRecord.SetIsDeployedBehavior(
							deploymentManifestPath,
							fakeCPIRelease,
							expectedExtractedStemcell,
							false,
							nil,
						)

						fakeDeploymentRecord.SetUpdateBehavior(
							deploymentManifestPath,
							fakeCPIRelease,
							nil,
						)
					})

					It("adds a new event logger stage", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())

						Expect(fakeEventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
							{
								Name: "validating",
							},
						}))

						Expect(fakeStage.Started).To(BeTrue())
						Expect(fakeStage.Finished).To(BeTrue())
					})

					It("parses the CPI portion of the manifest", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeCpiDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
					})

					It("parses the Bosh portion of the manifest", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeBoshManifestParser.ParseInputs[0].DeploymentPath).To(Equal(deploymentManifestPath))
					})

					It("validates bosh deployment manifest", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeDeploymentValidator.ValidateInputs).To(Equal([]fakebmdeplval.ValidateInput{
							{
								Deployment: boshDeployment,
							},
						}))
					})

					It("logs validation stages", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Validating deployment manifest",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Finished,
							},
						}))
						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Validating cpi release",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Finished,
							},
						}))
						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Validating stemcell",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Finished,
							},
						}))
					})

					It("extracts CPI release tarball", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeCPIInstaller.ExtractInputs).To(Equal([]fakebmcpi.ExtractInput{
							{
								ReleaseTarballPath: cpiReleaseTarballPath,
							},
						}))
					})

					It("installs the CPI locally", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeCPIInstaller.InstallInputs).To(Equal([]fakebmcpi.InstallInput{
							{
								Deployment: cpiDeployment,
								Release:    fakeCPIRelease,
							},
						}))
					})

					It("deletes the extracted CPI release", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeCPIRelease.DeleteCalled).To(BeTrue())
					})

					It("extracts the stemcell", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeStemcellExtractor.ExtractInputs).To(Equal([]fakebmstemcell.ExtractInput{
							{
								TarballPath: stemcellTarballPath,
							},
						}))
					})

					It("creates a VM", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeDeployer.DeployInputs).To(Equal([]fakebmdeployer.DeployInput{
							{
								Cpi:             cloud,
								Deployment:      boshDeployment,
								Stemcell:        expectedExtractedStemcell,
								Registry:        cpiDeployment.Registry,
								SSHTunnelConfig: cpiDeployment.SSHTunnel,
								MbusURL:         cpiDeployment.Mbus,
							},
						}))
					})

					It("updates the deployment record", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).NotTo(HaveOccurred())
						Expect(fakeDeploymentRecord.UpdateInputs).To(Equal([]fakebmdeployer.UpdateInput{
							{
								ManifestPath: deploymentManifestPath,
								Release:      fakeCPIRelease,
							},
						}))
					})

					Context("when deployment has not changed", func() {
						BeforeEach(func() {
							fakeDeploymentRecord.SetIsDeployedBehavior(
								deploymentManifestPath,
								fakeCPIRelease,
								expectedExtractedStemcell,
								true,
								nil,
							)
						})

						It("skips deploy", func() {
							err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
							Expect(err).NotTo(HaveOccurred())
							Expect(fakeUI.Said).To(ContainElement("No deployment, stemcell or cpi release changes. Skipping deploy."))
							Expect(fakeDeployer.DeployInputs).To(BeEmpty())
						})
					})

					Context("when parsing the cpi deployment manifest fails", func() {
						It("returns error", func() {
							fakeCpiDeploymentParser.ParseErr = errors.New("fake-parse-error")

							err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Parsing CPI deployment manifest"))
							Expect(err.Error()).To(ContainSubstring("fake-parse-error"))
							Expect(fakeCpiDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
						})
					})

					Context("when deployment validation fails", func() {
						BeforeEach(func() {
							fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
								{
									Err: errors.New("fake-validation-error"),
								},
							})
						})

						It("returns an error", func() {
							err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("fake-validation-error"))
						})

						It("logs the failed event log", func() {
							err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
							Expect(err).To(HaveOccurred())

							Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
								Name: "Validating deployment manifest",
								States: []bmeventlog.EventState{
									bmeventlog.Started,
									bmeventlog.Failed,
								},
								FailMessage: "Validating deployment manifest: fake-validation-error",
							}))
						})
					})

					Context("When the CPI release tarball does not exist", func() {
						BeforeEach(func() {
							fakeFs.RemoveAll(cpiReleaseTarballPath)
						})

						It("returns error", func() {
							err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Verifying that the CPI release `/release/tarball/path' exists"))

							Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
								Name: "Validating cpi release",
								States: []bmeventlog.EventState{
									bmeventlog.Started,
									bmeventlog.Failed,
								},
								FailMessage: "Verifying that the CPI release `/release/tarball/path' exists",
							}))
						})
					})

					Context("When the CPI stemcell tarball does not exist", func() {
						BeforeEach(func() {
							fakeFs.RemoveAll(stemcellTarballPath)
						})

						It("returns error", func() {
							err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Verifying that the stemcell `/stemcell/tarball/path' exists"))

							Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
								Name: "Validating stemcell",
								States: []bmeventlog.EventState{
									bmeventlog.Started,
									bmeventlog.Failed,
								},
								FailMessage: "Verifying that the stemcell `/stemcell/tarball/path' exists",
							}))
						})
					})
				})

				Context("when the deployment manifest is missing", func() {
					BeforeEach(func() {
						fakeFs.RemoveAll(deploymentManifestPath)
					})

					It("returns err", func() {
						err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Verifying that the deployment `/some/deployment/file' exists"))

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Validating deployment manifest",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Failed,
							},
							FailMessage: "Verifying that the deployment `/some/deployment/file' exists",
						}))
					})
				})
			})

			Context("when there is no deployment set", func() {
				BeforeEach(func() {
					userConfig.DeploymentFile = ""

					// re-create command to update userConfig.DeploymentFile
					command = bmcmd.NewDeployCmd(
						fakeUI,
						userConfig,
						fakeFs,
						fakeCpiDeploymentParser,
						fakeBoshManifestParser,
						fakeDeploymentValidator,
						fakeCPIInstaller,
						fakeStemcellExtractor,
						fakeDeploymentRecord,
						fakeDeployer,
						fakeEventLogger,
						logger,
					)
				})

				It("returns err", func() {
					err := command.Run([]string{cpiReleaseTarballPath, stemcellTarballPath})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("No deployment set"))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Validating deployment manifest",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Failed,
						},
						FailMessage: "No deployment set",
					}))
				})
			})
		})
	})
})
