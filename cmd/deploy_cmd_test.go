package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_cpi "github.com/cloudfoundry/bosh-micro-cli/cpi/mocks"
	mock_httpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"
	mock_deployer "github.com/cloudfoundry/bosh-micro-cli/deployment/mocks"
	mock_vm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/mocks"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"
	mock_release "github.com/cloudfoundry/bosh-micro-cli/release/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/fakes"
	fakebmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/validator/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		command    bmcmd.Cmd
		userConfig bmconfig.UserConfig
		fakeFs     *fakesys.FakeFileSystem
		fakeUI     *fakeui.FakeUI

		mockDeploymentFactory     *mock_deployer.MockFactory
		mockCPIDeploymentFactory  *mock_cpi.MockDeploymentFactory
		mockReleaseManager        *mock_release.MockManager
		mockRegistryServerManager *mock_registry.MockServerManager
		mockRegistryServer        *mock_registry.MockServer
		mockAgentClient           *mock_agentclient.MockAgentClient
		mockAgentClientFactory    *mock_httpagent.MockAgentClientFactory

		fakeCPIInstaller      *fakebmcpi.FakeInstaller
		fakeCPIRelease        *fakebmrel.FakeRelease
		logger                boshlog.Logger
		mockVMManagerFactory  *mock_vm.MockManagerFactory
		fakeVMManager         *fakebmvm.FakeManager
		fakeStemcellExtractor *fakebmstemcell.FakeExtractor

		fakeDeployer         *fakebmdepl.FakeDeployer
		fakeDeploymentRecord *fakebmdepl.FakeDeploymentRecord

		fakeDeploymentParser    *fakebmmanifest.FakeParser
		deploymentConfigService bmconfig.DeploymentConfigService
		fakeDeploymentValidator *fakebmdeplval.FakeValidator

		fakeCompressor    *fakecmd.FakeCompressor
		fakeJobRenderer   *fakebmtemp.FakeJobRenderer
		fakeUUIDGenerator *fakeuuid.FakeGenerator

		fakeEventLogger *fakebmlog.FakeEventLogger
		fakeStage       *fakebmlog.FakeStage

		deploymentManifestPath    string
		deploymentConfigPath      string
		cpiReleaseTarballPath     string
		stemcellTarballPath       string
		expectedExtractedStemcell bmstemcell.ExtractedStemcell
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		deploymentManifestPath = "/path/to/manifest.yml"
		deploymentConfigPath = "/path/to/deployment.json"
		userConfig = bmconfig.UserConfig{
			DeploymentManifestPath: deploymentManifestPath,
		}
		fakeFs.WriteFileString(deploymentManifestPath, "")

		mockDeploymentFactory = mock_deployer.NewMockFactory(mockCtrl)
		mockCPIDeploymentFactory = mock_cpi.NewMockDeploymentFactory(mockCtrl)

		mockReleaseManager = mock_release.NewMockManager(mockCtrl)

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)
		mockRegistryServer = mock_registry.NewMockServer(mockCtrl)

		mockAgentClientFactory = mock_httpagent.NewMockAgentClientFactory(mockCtrl)
		mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)
		mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), gomock.Any()).Return(mockAgentClient).AnyTimes()

		fakeCPIInstaller = fakebmcpi.NewFakeInstaller()

		mockVMManagerFactory = mock_vm.NewMockManagerFactory(mockCtrl)
		fakeVMManager = fakebmvm.NewFakeManager()
		mockVMManagerFactory.EXPECT().NewManager(gomock.Any(), mockAgentClient, gomock.Any()).Return(fakeVMManager).AnyTimes()

		fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()

		fakeDeployer = fakebmdepl.NewFakeDeployer()

		fakeDeploymentParser = fakebmmanifest.NewFakeParser()

		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, fakeFs, fakeUUIDGenerator, logger)

		fakeDeploymentValidator = fakebmdeplval.NewFakeValidator()

		fakeEventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior(fakeStage)

		fakeCompressor = fakecmd.NewFakeCompressor()
		fakeJobRenderer = fakebmtemp.NewFakeJobRenderer()

		fakeDeploymentRecord = fakebmdepl.NewFakeDeploymentRecord()

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

	JustBeforeEach(func() {
		command = bmcmd.NewDeployCmd(
			fakeUI,
			userConfig,
			fakeFs,
			fakeDeploymentParser,
			deploymentConfigService,
			fakeDeploymentValidator,
			mockCPIDeploymentFactory,
			mockReleaseManager,
			mockAgentClientFactory,
			mockVMManagerFactory,
			fakeStemcellExtractor,
			fakeDeploymentRecord,
			mockDeploymentFactory,
			fakeEventLogger,
			logger,
		)
	})

	Describe("Run", func() {
		var (
			boshDeploymentManifest bmmanifest.Manifest
			cpiDeploymentManifest  bmmanifest.CPIDeploymentManifest
			cloud                  *fakebmcloud.FakeCloud

			directorID   = "fake-uuid-0"
			deploymentID = "fake-uuid-1"

			expectCPIReleaseExtract *gomock.Call
		)

		BeforeEach(func() {
			// create input files
			fakeFs.WriteFileString(cpiReleaseTarballPath, "")
			fakeFs.WriteFileString(stemcellTarballPath, "")

			// deployment is set
			userConfig.DeploymentManifestPath = deploymentManifestPath

			// deployment exists
			fakeFs.WriteFileString(userConfig.DeploymentManifestPath, "")

			// deployment is valid
			fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
				{Err: nil},
			})

			// stemcell exists
			fakeFs.WriteFile(stemcellTarballPath, []byte{})

			// parsed CPI deployment manifest
			cpiDeploymentManifest = bmmanifest.CPIDeploymentManifest{
				Registry: bmmanifest.Registry{},
				SSHTunnel: bmmanifest.SSHTunnel{
					Host: "fake-host",
				},
				Mbus: "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint",
			}

			// parsed BOSH deployment manifest
			boshDeploymentManifest = bmmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bmmanifest.Job{
					{
						Name: "fake-job-name",
					},
				},
			}
			fakeDeploymentParser.ParseDeployment = boshDeploymentManifest

			// parsed/extracted CPI release
			fakeCPIRelease = fakebmrel.NewFakeRelease()
			fakeCPIRelease.ReleaseJobs = []bmrel.Job{
				{
					Name: "cpi",
					Templates: map[string]string{
						"templates/cpi.erb": "bin/cpi",
					},
				},
			}

			cloud = fakebmcloud.NewFakeCloud()
		})

		// allow return values of mocked methods to be modified by BeforeEach in child contexts
		JustBeforeEach(func() {
			fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, expectedExtractedStemcell, nil)

			fakeDeploymentParser.ParseDeployment = boshDeploymentManifest
			fakeDeploymentParser.ParseCPIDeploymentManifest = cpiDeploymentManifest

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

			cpiDeployment := bmcpi.NewDeployment(cpiDeploymentManifest, mockRegistryServerManager, fakeCPIInstaller, directorID)
			mockCPIDeploymentFactory.EXPECT().NewDeployment(cpiDeploymentManifest, deploymentID, directorID).Return(cpiDeployment).AnyTimes()

			deployment := bmdepl.NewDeployment(boshDeploymentManifest, fakeDeployer)
			mockDeploymentFactory.EXPECT().NewDeployment(boshDeploymentManifest).Return(deployment).AnyTimes()

			expectCPIReleaseExtract = mockReleaseManager.EXPECT().Extract(cpiReleaseTarballPath).Return(fakeCPIRelease, nil).AnyTimes()
			mockReleaseManager.EXPECT().List().Return([]bmrel.Release{fakeCPIRelease}).AnyTimes()
			mockReleaseManager.EXPECT().DeleteAll().Do(func() {
				err := fakeCPIRelease.Delete()
				Expect(err).ToNot(HaveOccurred())
			}).AnyTimes()

			fakeCPIInstaller.SetInstallBehavior(cpiDeploymentManifest, directorID, cloud, nil)

			fakeDeployer.SetDeployBehavior(nil)
		})

		It("prints the deployment manifest and state file", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeUI.Said).To(Equal([]string{
				"Deployment manifest: '/path/to/manifest.yml'",
				"Deployment state: '/path/to/deployment.json'",
			}))
		})

		It("adds a new event logger stage", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
				{Name: "validating"},
			}))

			Expect(fakeStage.Started).To(BeTrue())
			Expect(fakeStage.Finished).To(BeTrue())
		})

		It("parses the deployment manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("validates bosh deployment manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentValidator.ValidateInputs).To(Equal([]fakebmdeplval.ValidateInput{
				{Deployment: boshDeploymentManifest},
			}))
		})

		It("logs validation stages", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(Equal([]*fakebmlog.FakeStep{
				&fakebmlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmlog.FakeStep{
					Name: "Validating stemcell",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmlog.FakeStep{
					Name: "Validating releases",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
			}))
		})

		It("extracts CPI release tarball", func() {
			expectCPIReleaseExtract.Times(1)

			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		It("installs the CPI locally", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCPIInstaller.InstallInputs).To(Equal([]fakebmcpi.InstallInput{
				{
					Deployment: cpiDeploymentManifest,
					DirectorID: directorID,
				},
			}))
		})

		Context("when the registry is configured", func() {
			BeforeEach(func() {
				cpiDeploymentManifest.Registry = bmmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
			})

			It("starts & stops the registry", func() {
				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)
				mockRegistryServer.EXPECT().Stop()

				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("deletes the extracted CPI release", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCPIRelease.DeleteCalled).To(BeTrue())
		})

		It("extracts the stemcell", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeStemcellExtractor.ExtractInputs).To(Equal([]fakebmstemcell.ExtractInput{
				{TarballPath: stemcellTarballPath},
			}))
		})

		It("creates a VM", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeployer.DeployInputs).To(Equal([]fakebmdepl.DeployInput{
				{
					Cpi:             cloud,
					Manifest:        boshDeploymentManifest,
					Stemcell:        expectedExtractedStemcell,
					Registry:        cpiDeploymentManifest.Registry,
					SSHTunnelConfig: cpiDeploymentManifest.SSHTunnel,
					VMManager:       fakeVMManager,
				},
			}))
		})

		It("updates the deployment record", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentRecord.UpdateInputs).To(Equal([]fakebmdepl.UpdateInput{
				{
					ManifestPath: deploymentManifestPath,
					Release:      fakeCPIRelease,
				},
			}))
		})

		Context("when deployment has not changed", func() {
			JustBeforeEach(func() {
				fakeDeploymentRecord.SetIsDeployedBehavior(
					deploymentManifestPath,
					fakeCPIRelease,
					expectedExtractedStemcell,
					true,
					nil,
				)
			})

			It("skips deploy", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUI.Said).To(ContainElement("No deployment, stemcell or cpi release changes. Skipping deploy."))
				Expect(fakeDeployer.DeployInputs).To(BeEmpty())
			})
		})

		Context("when parsing the cpi deployment manifest fails", func() {
			BeforeEach(func() {
				fakeDeploymentParser.ParseErr = errors.New("fake-parse-error")
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing deployment manifest"))
				Expect(err.Error()).To(ContainSubstring("fake-parse-error"))
				Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
			})
		})

		Context("when the cpi release does not contain a 'cpi' job", func() {
			BeforeEach(func() {
				fakeCPIRelease.ReleaseJobs = []bmrel.Job{
					{
						Name: "not-cpi",
					},
				}
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("No provided release contains the required 'cpi' job"))
			})
		})

		Context("when multiple releases are given", func() {
			var (
				otherReleaseTarballPath string
				fakeOtherRelease        *fakebmrel.FakeRelease

				expectOtherReleaseExtract *gomock.Call
			)
			BeforeEach(func() {
				otherReleaseTarballPath = "/path/to/other-release.tgz"

				fakeFs.WriteFileString(otherReleaseTarballPath, "")

				fakeOtherRelease = fakebmrel.NewFakeRelease()
				fakeOtherRelease.ReleaseJobs = []bmrel.Job{
					{
						Name: "not-cpi",
					},
				}

				expectOtherReleaseExtract = mockReleaseManager.EXPECT().Extract(otherReleaseTarballPath).Return(fakeOtherRelease, nil).AnyTimes()
			})

			It("extracts all the release tarballs", func() {
				expectCPIReleaseExtract.Times(1)
				expectOtherReleaseExtract.Times(1)

				err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			It("installs the CPI release locally", func() {
				err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPIInstaller.InstallInputs).To(Equal([]fakebmcpi.InstallInput{
					{
						Deployment: cpiDeploymentManifest,
						DirectorID: directorID,
					},
				}))
			})

			//TODO: test all the normal deploy behavior (with multiple releases)?

			Context("when none of the releases include a 'cpi' job", func() {
				BeforeEach(func() {
					fakeCPIRelease.ReleaseJobs = []bmrel.Job{
						{
							Name: "not-cpi",
						},
					}
				})

				It("returns error", func() {
					err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("No provided release contains the required 'cpi' job"))
				})
			})
		})

		Context("When the CPI release tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(cpiReleaseTarballPath)
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the release '/release/tarball/path' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating releases",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the release '/release/tarball/path' exists",
				}))
			})
		})

		Context("When the stemcell tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(stemcellTarballPath)
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the stemcell '/stemcell/tarball/path' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating stemcell",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the stemcell '/stemcell/tarball/path' exists",
				}))
			})
		})

		Context("when the deployment config file does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(deploymentConfigPath)
			})

			It("creates a deployment config", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := deploymentConfigService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig).To(Equal(bmconfig.DeploymentFile{
					DirectorID:   "fake-uuid-0",
					DeploymentID: "fake-uuid-1",
				}))
			})
		})

		Context("when there is no deployment set", func() {
			BeforeEach(func() {
				userConfig.DeploymentManifestPath = ""

				// re-create command to update userConfig.DeploymentFile
				command = bmcmd.NewDeployCmd(
					fakeUI,
					userConfig,
					fakeFs,
					fakeDeploymentParser,
					deploymentConfigService,
					fakeDeploymentValidator,
					mockCPIDeploymentFactory,
					mockReleaseManager,
					mockAgentClientFactory,
					mockVMManagerFactory,
					fakeStemcellExtractor,
					fakeDeploymentRecord,
					mockDeploymentFactory,
					fakeEventLogger,
					logger,
				)
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Running deploy cmd: Deployment manifest not set"))
				Expect(fakeUI.Errors).To(ContainElement("Deployment manifest not set"))
			})
		})

		Context("when the deployment manifest does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(deploymentManifestPath)
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Running deploy cmd: Deployment manifest does not exist at '/path/to/manifest.yml'"))
				Expect(fakeUI.Errors).To(ContainElement("Deployment manifest does not exist"))
			})
		})

		Context("when the deployment manifest is invalid", func() {
			BeforeEach(func() {
				fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
					{Err: errors.New("fake-deployment-validation-error")},
				})
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-deployment-validation-error"))
			})

			It("logs the failed event log", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Validating deployment manifest: fake-deployment-validation-error",
				}))
			})
		})

		It("returns err when no arguments are given", func() {
			err := command.Run([]string{})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires at least 2 arguments"))
		})

		It("returns err when 1 argument is given", func() {
			err := command.Run([]string{"something"})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires at least 2 arguments"))
		})
	})
})
