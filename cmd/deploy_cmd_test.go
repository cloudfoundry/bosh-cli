package cmd_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_cloud "github.com/cloudfoundry/bosh-micro-cli/cloud/mocks"
	mock_httpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"
	mock_deployment "github.com/cloudfoundry/bosh-micro-cli/deployment/mocks"
	mock_vm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/mocks"
	mock_install "github.com/cloudfoundry/bosh-micro-cli/installation/mocks"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"
	mock_release "github.com/cloudfoundry/bosh-micro-cli/release/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/fakes"
	fakebmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakebmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"
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

		mockDeployer              *mock_deployment.MockDeployer
		mockInstaller             *mock_install.MockInstaller
		mockInstallerFactory      *mock_install.MockInstallerFactory
		mockReleaseExtractor      *mock_release.MockExtractor
		releaseManager            bmrel.Manager
		releaseSetResolver        bmrelset.Resolver
		mockRegistryServerManager *mock_registry.MockServerManager
		mockRegistryServer        *mock_registry.MockServer
		mockAgentClient           *mock_agentclient.MockAgentClient
		mockAgentClientFactory    *mock_httpagent.MockAgentClientFactory
		mockCloudFactory          *mock_cloud.MockFactory

		fakeCPIRelease *fakebmrel.FakeRelease
		logger         boshlog.Logger

		mockBlobstoreFactory *mock_blobstore.MockFactory
		mockBlobstore        *mock_blobstore.MockBlobstore

		mockVMManagerFactory       *mock_vm.MockManagerFactory
		fakeVMManager              *fakebmvm.FakeManager
		fakeStemcellExtractor      *fakebmstemcell.FakeExtractor
		fakeStemcellManager        *fakebmstemcell.FakeManager
		fakeStemcellManagerFactory *fakebmstemcell.FakeManagerFactory

		fakeDeploymentRecord *fakebmdepl.FakeRecord

		fakeReleaseSetParser      *fakebmrelsetmanifest.FakeParser
		fakeInstallationParser    *fakebminstallmanifest.FakeParser
		fakeDeploymentParser      *fakebmdeplmanifest.FakeParser
		deploymentConfigService   bmconfig.DeploymentConfigService
		fakeReleaseSetValidator   *fakebmrelsetmanifest.FakeValidator
		fakeInstallationValidator *fakebminstallmanifest.FakeValidator
		fakeDeploymentValidator   *fakebmdeplval.FakeValidator

		fakeUUIDGenerator *fakeuuid.FakeGenerator

		fakeEventLogger     *fakebmeventlog.FakeEventLogger
		fakeValidatingStage *fakebmeventlog.FakeStage
		fakeInstallingStage *fakebmeventlog.FakeStage
		fakeUploadingStage  *fakebmeventlog.FakeStage
		fakeDeployingStage  *fakebmeventlog.FakeStage

		deploymentManifestPath string
		deploymentConfigPath   string
		cpiReleaseTarballPath  string
		stemcellTarballPath    string
		extractedStemcell      bmstemcell.ExtractedStemcell

		expectDeploy *gomock.Call

		mbusURL = "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint"
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		deploymentManifestPath = "/path/to/manifest.yml"
		deploymentConfigPath = "/path/to/deployment.json"
		userConfig = bmconfig.UserConfig{
			DeploymentManifestPath: deploymentManifestPath,
		}
		fakeFs.WriteFileString(deploymentManifestPath, "")

		mockDeployer = mock_deployment.NewMockDeployer(mockCtrl)
		mockInstaller = mock_install.NewMockInstaller(mockCtrl)
		mockInstallerFactory = mock_install.NewMockInstallerFactory(mockCtrl)

		mockReleaseExtractor = mock_release.NewMockExtractor(mockCtrl)
		releaseManager = bmrel.NewManager(logger)
		releaseSetResolver = bmrelset.NewResolver(releaseManager, logger)

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)
		mockRegistryServer = mock_registry.NewMockServer(mockCtrl)

		mockAgentClientFactory = mock_httpagent.NewMockAgentClientFactory(mockCtrl)
		mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)
		mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), gomock.Any()).Return(mockAgentClient).AnyTimes()

		mockCloudFactory = mock_cloud.NewMockFactory(mockCtrl)

		mockBlobstoreFactory = mock_blobstore.NewMockFactory(mockCtrl)
		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)
		mockBlobstoreFactory.EXPECT().Create(mbusURL).Return(mockBlobstore, nil).AnyTimes()

		mockVMManagerFactory = mock_vm.NewMockManagerFactory(mockCtrl)
		fakeVMManager = fakebmvm.NewFakeManager()
		mockVMManagerFactory.EXPECT().NewManager(gomock.Any(), mockAgentClient).Return(fakeVMManager).AnyTimes()

		fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()
		fakeStemcellManager = fakebmstemcell.NewFakeManager()
		fakeStemcellManagerFactory = fakebmstemcell.NewFakeManagerFactory()

		fakeReleaseSetParser = fakebmrelsetmanifest.NewFakeParser()
		fakeInstallationParser = fakebminstallmanifest.NewFakeParser()
		fakeDeploymentParser = fakebmdeplmanifest.NewFakeParser()

		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, fakeFs, fakeUUIDGenerator, logger)

		fakeReleaseSetValidator = fakebmrelsetmanifest.NewFakeValidator()
		fakeInstallationValidator = fakebminstallmanifest.NewFakeValidator()
		fakeDeploymentValidator = fakebmdeplval.NewFakeValidator()

		fakeEventLogger = fakebmeventlog.NewFakeEventLogger()
		fakeValidatingStage = fakebmeventlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior("validating", fakeValidatingStage)
		fakeInstallingStage = fakebmeventlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior("installing CPI", fakeInstallingStage)
		fakeUploadingStage = fakebmeventlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior("uploading stemcell", fakeUploadingStage)
		fakeDeployingStage = fakebmeventlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior("deploying", fakeDeployingStage)

		fakeDeploymentRecord = fakebmdepl.NewFakeRecord()

		cpiReleaseTarballPath = "/release/tarball/path"

		stemcellTarballPath = "/stemcell/tarball/path"
		extractedStemcell = bmstemcell.NewExtractedStemcell(
			bmstemcell.Manifest{
				ImagePath:       "/stemcell/image/path",
				Name:            "fake-stemcell-name",
				Version:         "fake-stemcell-version",
				SHA1:            "fake-stemcell-sha1",
				CloudProperties: bmproperty.Map{},
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
			fakeReleaseSetParser,
			fakeInstallationParser,
			fakeDeploymentParser,
			deploymentConfigService,
			fakeReleaseSetValidator,
			fakeInstallationValidator,
			fakeDeploymentValidator,
			mockInstallerFactory,
			mockReleaseExtractor,
			releaseManager,
			releaseSetResolver,
			mockCloudFactory,
			mockAgentClientFactory,
			mockVMManagerFactory,
			fakeStemcellExtractor,
			fakeStemcellManagerFactory,
			fakeDeploymentRecord,
			mockBlobstoreFactory,
			mockDeployer,
			fakeEventLogger,
			logger,
		)
	})

	Describe("Run", func() {
		var (
			releaseSetManifest     bmrelsetmanifest.Manifest
			boshDeploymentManifest bmdeplmanifest.Manifest
			installationManifest   bminstallmanifest.Manifest
			fakeCloud              *fakebmcloud.FakeCloud

			cloudStemcell bmstemcell.CloudStemcell

			directorID = "fake-uuid-0"

			expectCPIReleaseExtract *gomock.Call
			expectInstall           *gomock.Call
			expectNewCloud          *gomock.Call
		)

		BeforeEach(func() {
			// create input files
			fakeFs.WriteFileString(cpiReleaseTarballPath, "")
			fakeFs.WriteFileString(stemcellTarballPath, "")

			// deployment is set
			userConfig.DeploymentManifestPath = deploymentManifestPath

			// deployment exists
			fakeFs.WriteFileString(userConfig.DeploymentManifestPath, "")

			// release set is valid
			fakeReleaseSetValidator.SetValidateBehavior([]fakebmrelsetmanifest.ValidateOutput{
				{Err: nil},
			})

			fakeInstallationValidator.SetValidateBehavior([]fakebminstallmanifest.ValidateOutput{
				{Err: nil},
			})

			// deployment is valid
			fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
				{Err: nil},
			})

			// stemcell exists
			fakeFs.WriteFile(stemcellTarballPath, []byte{})

			releaseSetManifest = bmrelsetmanifest.Manifest{
				Releases: []bmrelmanifest.ReleaseRef{
					{
						Name:    "fake-cpi-release-name",
						Version: "1.0",
					},
				},
			}

			// parsed CPI deployment manifest
			installationManifest = bminstallmanifest.Manifest{
				Release:  "fake-cpi-release-name",
				Registry: bminstallmanifest.Registry{},
				SSHTunnel: bminstallmanifest.SSHTunnel{
					Host: "fake-host",
				},
				Mbus: mbusURL,
			}

			// parsed BOSH deployment manifest
			boshDeploymentManifest = bmdeplmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bmdeplmanifest.Job{
					{
						Name: "fake-job-name",
					},
				},
			}
			fakeDeploymentParser.ParseManifest = boshDeploymentManifest

			// parsed/extracted CPI release
			fakeCPIRelease = fakebmrel.NewFakeRelease()
			fakeCPIRelease.ReleaseName = "fake-cpi-release-name"
			fakeCPIRelease.ReleaseVersion = "1.0"
			fakeCPIRelease.ReleaseJobs = []bmreljob.Job{
				{
					Name: "cpi",
					Templates: map[string]string{
						"templates/cpi.erb": "bin/cpi",
					},
				},
			}

			fakeCloud = fakebmcloud.NewFakeCloud()

			cloudStemcell = fakebmstemcell.NewFakeCloudStemcell("fake-stemcell-cid", "fake-stemcell-name", "fake-stemcell-version")
		})

		// allow return values of mocked methods to be modified by BeforeEach in child contexts
		JustBeforeEach(func() {
			fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, extractedStemcell, nil)

			fakeStemcellManagerFactory.SetNewManagerBehavior(fakeCloud, fakeStemcellManager)
			fakeStemcellManager.SetUploadBehavior(extractedStemcell, fakeUploadingStage, cloudStemcell, nil)

			fakeReleaseSetParser.ParseManifest = releaseSetManifest
			fakeDeploymentParser.ParseManifest = boshDeploymentManifest
			fakeInstallationParser.ParseManifest = installationManifest

			fakeDeploymentRecord.SetIsDeployedBehavior(
				deploymentManifestPath,
				fakeCPIRelease,
				extractedStemcell,
				false,
				nil,
			)

			fakeDeploymentRecord.SetUpdateBehavior(
				deploymentManifestPath,
				fakeCPIRelease,
				nil,
			)

			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := bminstall.NewTarget(installationPath)

			installedJob := bminstalljob.InstalledJob{
				Name: "cpi",
				Path: filepath.Join(target.JobsPath(), "cpi"),
			}

			mockInstallerFactory.EXPECT().NewInstaller().Return(mockInstaller, nil).AnyTimes()

			installation := bminstall.NewInstallation(target, installedJob, installationManifest, mockRegistryServerManager)

			expectInstall = mockInstaller.EXPECT().Install(installationManifest, fakeInstallingStage).Return(installation, nil).AnyTimes()

			mockDeployment := mock_deployment.NewMockDeployment(mockCtrl)

			expectDeploy = mockDeployer.EXPECT().Deploy(
				fakeCloud,
				boshDeploymentManifest,
				cloudStemcell,
				installationManifest.Registry,
				installationManifest.SSHTunnel,
				fakeVMManager,
				mockBlobstore,
				fakeDeployingStage,
			).Return(mockDeployment, nil).AnyTimes()

			expectCPIReleaseExtract = mockReleaseExtractor.EXPECT().Extract(cpiReleaseTarballPath).Return(fakeCPIRelease, nil).AnyTimes()

			expectNewCloud = mockCloudFactory.EXPECT().NewCloud(installation, directorID).Return(fakeCloud, nil).AnyTimes()
		})

		It("prints the deployment manifest and state file", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeUI.Said).To(Equal([]string{
				"Deployment manifest: '/path/to/manifest.yml'",
				"Deployment state: '/path/to/deployment.json'",
			}))
		})

		It("adds a new 'validating' event logger stage", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEventLogger.NewStageInputs[0]).To(Equal(fakebmeventlog.NewStageInput{
				Name: "validating",
			}))

			Expect(fakeValidatingStage.Started).To(BeTrue())
			Expect(fakeValidatingStage.Finished).To(BeTrue())
		})

		It("parses the installation manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("parses the deployment manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("validates release set manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleaseSetValidator.ValidateInputs).To(Equal([]fakebmrelsetmanifest.ValidateInput{
				{Manifest: releaseSetManifest},
			}))
		})

		It("validates installation manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationValidator.ValidateInputs).To(Equal([]fakebminstallmanifest.ValidateInput{
				{Manifest: installationManifest},
			}))
		})

		It("validates bosh deployment manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentValidator.ValidateInputs).To(Equal([]fakebmdeplval.ValidateInput{
				{Manifest: boshDeploymentManifest},
			}))
		})

		It("logs validating stages", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeValidatingStage.Steps).To(Equal([]*fakebmeventlog.FakeStep{
				&fakebmeventlog.FakeStep{
					Name: "Validating stemcell",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmeventlog.FakeStep{
					Name: "Validating releases",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmeventlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmeventlog.FakeStep{
					Name: "Validating cpi release",
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
			expectInstall.Times(1)
			expectNewCloud.Times(1)

			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		It("adds a new 'installing CPI' event logger stage", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEventLogger.NewStageInputs[1]).To(Equal(fakebmeventlog.NewStageInput{
				Name: "installing CPI",
			}))

			Expect(fakeInstallingStage.Started).To(BeTrue())
			Expect(fakeInstallingStage.Finished).To(BeTrue())
		})

		Context("when the registry is configured", func() {
			BeforeEach(func() {
				installationManifest.Registry = bminstallmanifest.Registry{
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

		It("adds a new 'uploading stemcell' event logger stage", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEventLogger.NewStageInputs[2]).To(Equal(fakebmeventlog.NewStageInput{
				Name: "uploading stemcell",
			}))

			Expect(fakeUploadingStage.Started).To(BeTrue())
			Expect(fakeUploadingStage.Finished).To(BeTrue())
		})

		It("uploads the stemcell", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeStemcellManager.UploadInputs).To(Equal([]fakebmstemcell.UploadInput{
				{Stemcell: extractedStemcell, Stage: fakeUploadingStage},
			}))
		})

		It("adds a new 'deploying' event logger stage", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEventLogger.NewStageInputs[3]).To(Equal(fakebmeventlog.NewStageInput{
				Name: "deploying",
			}))

			Expect(fakeDeployingStage.Started).To(BeTrue())
			Expect(fakeDeployingStage.Finished).To(BeTrue())
		})

		It("deploys", func() {
			expectDeploy.Times(1)

			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
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

		It("deletes unused stemcells", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStemcellManager.DeleteUnusedCalledTimes).To(Equal(1))
		})

		Context("when deployment has not changed", func() {
			JustBeforeEach(func() {
				fakeDeploymentRecord.SetIsDeployedBehavior(
					deploymentManifestPath,
					fakeCPIRelease,
					extractedStemcell,
					true,
					nil,
				)
			})

			It("skips deploy", func() {
				expectDeploy.Times(0)

				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUI.Said).To(ContainElement("No deployment, stemcell or cpi release changes. Skipping deploy."))
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
				fakeCPIRelease.ReleaseJobs = []bmreljob.Job{
					{
						Name: "not-cpi",
					},
				}
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid CPI release 'fake-cpi-release-name': Job 'cpi' is missing from release"))
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
				fakeOtherRelease.ReleaseJobs = []bmreljob.Job{
					{
						Name: "not-cpi",
					},
				}

				expectOtherReleaseExtract = mockReleaseExtractor.EXPECT().Extract(otherReleaseTarballPath).Return(fakeOtherRelease, nil).AnyTimes()
			})

			It("extracts all the release tarballs", func() {
				expectCPIReleaseExtract.Times(1)
				expectOtherReleaseExtract.Times(1)

				err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			It("installs the CPI release locally", func() {
				expectInstall.Times(1)
				expectNewCloud.Times(1)

				err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when cloud_provider.release refers to an release declared with version 'latest'", func() {
				BeforeEach(func() {
					releaseSetManifest.Releases = []bmrelmanifest.ReleaseRef{
						{
							Name:    "fake-cpi-release-name",
							Version: "latest",
						},
					}
				})

				It("uses the latest version of that release that is available", func() {
					expectInstall.Times(1)
					expectNewCloud.Times(1)

					err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
					Expect(err).NotTo(HaveOccurred())
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

				Expect(fakeValidatingStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
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

				Expect(fakeValidatingStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
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
					DirectorID: directorID,
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
					fakeReleaseSetParser,
					fakeInstallationParser,
					fakeDeploymentParser,
					deploymentConfigService,
					fakeReleaseSetValidator,
					fakeInstallationValidator,
					fakeDeploymentValidator,
					mockInstallerFactory,
					mockReleaseExtractor,
					releaseManager,
					releaseSetResolver,
					mockCloudFactory,
					mockAgentClientFactory,
					mockVMManagerFactory,
					fakeStemcellExtractor,
					fakeStemcellManagerFactory,
					fakeDeploymentRecord,
					mockBlobstoreFactory,
					mockDeployer,
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

		Context("when the installation manifest is invalid", func() {
			BeforeEach(func() {
				fakeInstallationValidator.SetValidateBehavior([]fakebminstallmanifest.ValidateOutput{
					{Err: errors.New("fake-installation-validation-error")},
				})
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-installation-validation-error"))
			})

			It("logs the failed event log", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())

				Expect(fakeValidatingStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Validating installation manifest: fake-installation-validation-error",
				}))
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

				Expect(fakeValidatingStage.Steps).To(ContainElement(&fakebmeventlog.FakeStep{
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

		Context("when uploading stemcell fails", func() {
			JustBeforeEach(func() {
				fakeStemcellManager.SetUploadBehavior(extractedStemcell, fakeUploadingStage, nil, errors.New("fake-upload-error"))
			})

			It("returns an error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-upload-error"))
			})
		})
	})
})
