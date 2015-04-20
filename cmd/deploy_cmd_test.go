package cmd

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-init/blobstore/mocks"
	mock_cloud "github.com/cloudfoundry/bosh-init/cloud/mocks"
	mock_config "github.com/cloudfoundry/bosh-init/config/mocks"
	mock_httpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient/mocks"
	mock_deployment "github.com/cloudfoundry/bosh-init/deployment/mocks"
	mock_vm "github.com/cloudfoundry/bosh-init/deployment/vm/mocks"
	mock_install "github.com/cloudfoundry/bosh-init/installation/mocks"
	mock_registry "github.com/cloudfoundry/bosh-init/registry/mocks"
	mock_release "github.com/cloudfoundry/bosh-init/release/mocks"
	mock_stemcell "github.com/cloudfoundry/bosh-init/stemcell/mocks"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	birel "github.com/cloudfoundry/bosh-init/release"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	biui "github.com/cloudfoundry/bosh-init/ui"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebicloud "github.com/cloudfoundry/bosh-init/cloud/fakes"
	fakebideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest/fakes"
	fakebideplval "github.com/cloudfoundry/bosh-init/deployment/manifest/fakes"
	fakebivm "github.com/cloudfoundry/bosh-init/deployment/vm/fakes"
	fakebiinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest/fakes"
	fakebirel "github.com/cloudfoundry/bosh-init/release/fakes"
	fakebirelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest/fakes"
	fakebistemcell "github.com/cloudfoundry/bosh-init/stemcell/fakes"
	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"

	"fmt"
	"github.com/cloudfoundry/bosh-init/crypto"
	"github.com/cloudfoundry/bosh-init/deployment"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DeployCmd", rootDesc)

func rootDesc() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Run", func() {
		var (
			command        Cmd
			fakeFs         *fakesys.FakeFileSystem
			stdOut         *gbytes.Buffer
			stdErr         *gbytes.Buffer
			userInterface  biui.UI
			sha1Calculator crypto.SHA1Calculator
			manifestSHA1   string

			mockDeployer              *mock_deployment.MockDeployer
			mockInstaller             *mock_install.MockInstaller
			mockInstallerFactory      *mock_install.MockInstallerFactory
			mockReleaseExtractor      *mock_release.MockExtractor
			releaseManager            birel.Manager
			mockRegistryServerManager *mock_registry.MockServerManager
			mockRegistryServer        *mock_registry.MockServer
			mockAgentClient           *mock_agentclient.MockAgentClient
			mockAgentClientFactory    *mock_httpagent.MockAgentClientFactory
			mockCloudFactory          *mock_cloud.MockFactory

			fakeCPIRelease *fakebirel.FakeRelease
			logger         boshlog.Logger

			mockBlobstoreFactory *mock_blobstore.MockFactory
			mockBlobstore        *mock_blobstore.MockBlobstore

			mockVMManagerFactory       *mock_vm.MockManagerFactory
			fakeVMManager              *fakebivm.FakeManager
			fakeStemcellExtractor      *fakebistemcell.FakeExtractor
			mockStemcellManager        *mock_stemcell.MockManager
			fakeStemcellManagerFactory *fakebistemcell.FakeManagerFactory

			fakeReleaseSetParser              *fakebirelsetmanifest.FakeParser
			fakeInstallationParser            *fakebiinstallmanifest.FakeParser
			fakeDeploymentParser              *fakebideplmanifest.FakeParser
			mockLegacyDeploymentStateMigrator *mock_config.MockLegacyDeploymentStateMigrator
			setupDeploymentStateService       biconfig.DeploymentStateService
			fakeReleaseSetValidator           *fakebirelsetmanifest.FakeValidator
			fakeInstallationValidator         *fakebiinstallmanifest.FakeValidator
			fakeDeploymentValidator           *fakebideplval.FakeValidator

			directorID          = "generated-director-uuid"
			fakeUUIDGenerator   *fakeuuid.FakeGenerator
			configUUIDGenerator *fakeuuid.FakeGenerator

			fakeStage *fakebiui.FakeStage

			deploymentManifestPath string
			deploymentStatePath    string
			cpiReleaseTarballPath  string
			stemcellTarballPath    string
			extractedStemcell      bistemcell.ExtractedStemcell

			expectDeploy *gomock.Call

			mbusURL = "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint"

			releaseSetManifest     birelsetmanifest.Manifest
			boshDeploymentManifest bideplmanifest.Manifest
			installationManifest   biinstallmanifest.Manifest
			cloud                  bicloud.Cloud
			deploymentReleases     []birel.Release

			cloudStemcell bistemcell.CloudStemcell

			expectLegacyMigrate        *gomock.Call
			expectStemcellUpload       *gomock.Call
			expectStemcellDeleteUnused *gomock.Call
			expectCPIReleaseExtract    *gomock.Call
			expectInstall              *gomock.Call
			expectNewCloud             *gomock.Call
		)

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)
			stdOut = gbytes.NewBuffer()
			stdErr = gbytes.NewBuffer()
			userInterface = biui.NewWriterUI(stdOut, stdErr, logger)
			fakeFs = fakesys.NewFakeFileSystem()
			deploymentManifestPath = "/path/to/manifest.yml"
			deploymentStatePath = "/path/to/manifest-state.json"
			fakeFs.RegisterOpenFile(deploymentManifestPath, &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
			})

			fakeFs.WriteFileString(deploymentManifestPath, "")

			mockDeployer = mock_deployment.NewMockDeployer(mockCtrl)
			mockInstaller = mock_install.NewMockInstaller(mockCtrl)
			mockInstallerFactory = mock_install.NewMockInstallerFactory(mockCtrl)

			mockReleaseExtractor = mock_release.NewMockExtractor(mockCtrl)
			releaseManager = birel.NewManager(logger)

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
			fakeVMManager = fakebivm.NewFakeManager()
			mockVMManagerFactory.EXPECT().NewManager(gomock.Any(), mockAgentClient).Return(fakeVMManager).AnyTimes()

			fakeStemcellExtractor = fakebistemcell.NewFakeExtractor()
			mockStemcellManager = mock_stemcell.NewMockManager(mockCtrl)
			fakeStemcellManagerFactory = fakebistemcell.NewFakeManagerFactory()

			fakeReleaseSetParser = fakebirelsetmanifest.NewFakeParser()
			fakeInstallationParser = fakebiinstallmanifest.NewFakeParser()
			fakeDeploymentParser = fakebideplmanifest.NewFakeParser()

			mockLegacyDeploymentStateMigrator = mock_config.NewMockLegacyDeploymentStateMigrator(mockCtrl)

			configUUIDGenerator = &fakeuuid.FakeGenerator{}
			configUUIDGenerator.GeneratedUUID = directorID
			setupDeploymentStateService = biconfig.NewFileSystemDeploymentStateService(fakeFs, configUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath))

			fakeReleaseSetValidator = fakebirelsetmanifest.NewFakeValidator()
			fakeInstallationValidator = fakebiinstallmanifest.NewFakeValidator()
			fakeDeploymentValidator = fakebideplval.NewFakeValidator()

			fakeStage = fakebiui.NewFakeStage()

			sha1Calculator = crypto.NewSha1Calculator(fakeFs)
			fakeUUIDGenerator = &fakeuuid.FakeGenerator{}

			var err error
			manifestSHA1, err = sha1Calculator.Calculate(deploymentManifestPath)
			Expect(err).ToNot(HaveOccurred())

			cpiReleaseTarballPath = "/release/tarball/path"

			stemcellTarballPath = "/stemcell/tarball/path"
			extractedStemcell = bistemcell.NewExtractedStemcell(
				bistemcell.Manifest{
					ImagePath:       "/stemcell/image/path",
					Name:            "fake-stemcell-name",
					Version:         "fake-stemcell-version",
					SHA1:            "fake-stemcell-sha1",
					CloudProperties: biproperty.Map{},
				},
				"fake-extracted-path",
				fakeFs,
			)

			// create input files
			fakeFs.WriteFileString(cpiReleaseTarballPath, "")
			fakeFs.WriteFileString(stemcellTarballPath, "")

			// deployment exists
			fakeFs.WriteFileString(deploymentManifestPath, "")

			// release set is valid
			fakeReleaseSetValidator.SetValidateBehavior([]fakebirelsetmanifest.ValidateOutput{
				{Err: nil},
			})

			fakeInstallationValidator.SetValidateBehavior([]fakebiinstallmanifest.ValidateOutput{
				{Err: nil},
			})

			// deployment is valid
			fakeDeploymentValidator.SetValidateBehavior([]fakebideplval.ValidateOutput{
				{Err: nil},
			})

			// stemcell exists
			fakeFs.WriteFile(stemcellTarballPath, []byte{})

			releaseSetManifest = birelsetmanifest.Manifest{
				Releases: []birelmanifest.ReleaseRef{
					{
						Name: "fake-cpi-release-name",
						URL:  "file://" + cpiReleaseTarballPath,
					},
				},
			}

			// parsed CPI deployment manifest
			installationManifest = biinstallmanifest.Manifest{
				Template: biinstallmanifest.ReleaseJobRef{
					Name:    "fake-cpi-release-job-name",
					Release: "fake-cpi-release-name",
				},
				Registry: biinstallmanifest.Registry{},
				SSHTunnel: biinstallmanifest.SSHTunnel{
					Host: "fake-host",
				},
				Mbus: mbusURL,
			}

			// parsed BOSH deployment manifest
			boshDeploymentManifest = bideplmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bideplmanifest.Job{
					{
						Name: "fake-job-name",
					},
				},
				ResourcePools: []bideplmanifest.ResourcePool{
					{
						Stemcell: bideplmanifest.StemcellRef{
							URL: "file://" + stemcellTarballPath,
						},
					},
				},
			}
			fakeDeploymentParser.ParseManifest = boshDeploymentManifest

			// parsed/extracted CPI release
			fakeCPIRelease = fakebirel.NewFakeRelease()
			deploymentReleases = []birel.Release{fakeCPIRelease}
			fakeCPIRelease.ReleaseName = "fake-cpi-release-name"
			fakeCPIRelease.ReleaseVersion = "1.0"
			fakeCPIRelease.ReleaseJobs = []bireljob.Job{
				{
					Name: "fake-cpi-release-job-name",
					Templates: map[string]string{
						"templates/cpi.erb": "bin/cpi",
					},
				},
			}

			cloud = bicloud.NewCloud(fakebicloud.NewFakeCPICmdRunner(), "fake-director-id", logger)

			cloudStemcell = fakebistemcell.NewFakeCloudStemcell("fake-stemcell-cid", "fake-stemcell-name", "fake-stemcell-version")
		})

		JustBeforeEach(func() {

			doGet := func(deploymentManifestPath string) DeploymentPreparer {
				deploymentStateService := biconfig.NewFileSystemDeploymentStateService(fakeFs, configUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath))
				deploymentRepo := biconfig.NewDeploymentRepo(deploymentStateService)
				releaseRepo := biconfig.NewReleaseRepo(deploymentStateService, fakeUUIDGenerator)
				stemcellRepo := biconfig.NewStemcellRepo(deploymentStateService, fakeUUIDGenerator)
				deploymentRecord := deployment.NewRecord(deploymentRepo, releaseRepo, stemcellRepo, sha1Calculator)
				return DeploymentPreparer{
					ui:     userInterface,
					fs:     fakeFs,
					logger: logger,
					logTag: "deployCmd",

					releaseSetParser:              fakeReleaseSetParser,
					installationParser:            fakeInstallationParser,
					deploymentParser:              fakeDeploymentParser,
					legacyDeploymentStateMigrator: mockLegacyDeploymentStateMigrator,
					deploymentStateService:        deploymentStateService,
					releaseSetValidator:           fakeReleaseSetValidator,
					installationValidator:         fakeInstallationValidator,
					deploymentValidator:           fakeDeploymentValidator,
					installerFactory:              mockInstallerFactory,
					releaseExtractor:              mockReleaseExtractor,
					releaseManager:                releaseManager,
					cloudFactory:                  mockCloudFactory,
					agentClientFactory:            mockAgentClientFactory,
					vmManagerFactory:              mockVMManagerFactory,
					stemcellExtractor:             fakeStemcellExtractor,
					stemcellManagerFactory:        fakeStemcellManagerFactory,
					deploymentRecord:              deploymentRecord,
					blobstoreFactory:              mockBlobstoreFactory,
					deployer:                      mockDeployer,
					deploymentManifestPath:        deploymentManifestPath,
				}
			}

			command = NewDeployCmd(userInterface, fakeFs, logger, doGet)

			expectLegacyMigrate = mockLegacyDeploymentStateMigrator.EXPECT().MigrateIfExists("/path/to/bosh-deployments.yml").AnyTimes()

			fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, extractedStemcell, nil)

			fakeStemcellManagerFactory.SetNewManagerBehavior(cloud, mockStemcellManager)

			expectStemcellUpload = mockStemcellManager.EXPECT().Upload(extractedStemcell, fakeStage).Return(cloudStemcell, nil).AnyTimes()

			expectStemcellDeleteUnused = mockStemcellManager.EXPECT().DeleteUnused(fakeStage).AnyTimes()

			fakeReleaseSetParser.ParseManifest = releaseSetManifest
			fakeDeploymentParser.ParseManifest = boshDeploymentManifest
			fakeInstallationParser.ParseManifest = installationManifest

			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := biinstall.NewTarget(installationPath)

			installedJob := biinstalljob.InstalledJob{
				Name: "fake-cpi-release-job-name",
				Path: filepath.Join(target.JobsPath(), "fake-cpi-release-job-name"),
			}

			mockInstallerFactory.EXPECT().NewInstaller().Return(mockInstaller, nil).AnyTimes()

			installation := biinstall.NewInstallation(target, installedJob, installationManifest, mockRegistryServerManager)

			expectInstall = mockInstaller.EXPECT().Install(installationManifest, gomock.Any()).Do(func(_ interface{}, stage biui.Stage) {
				Expect(fakeStage.SubStages).To(ContainElement(stage))
			}).Return(installation, nil).AnyTimes()

			mockDeployment := mock_deployment.NewMockDeployment(mockCtrl)

			expectDeploy = mockDeployer.EXPECT().Deploy(
				cloud,
				boshDeploymentManifest,
				cloudStemcell,
				installationManifest.Registry,
				installationManifest.SSHTunnel,
				fakeVMManager,
				mockBlobstore,
				gomock.Any(),
			).Do(func(_, _, _, _, _, _, _ interface{}, stage biui.Stage) {
				Expect(fakeStage.SubStages).To(ContainElement(stage))
			}).Return(mockDeployment, nil).AnyTimes()

			expectCPIReleaseExtract = mockReleaseExtractor.EXPECT().Extract(cpiReleaseTarballPath).Return(fakeCPIRelease, nil).AnyTimes()

			expectNewCloud = mockCloudFactory.EXPECT().NewCloud(installation, directorID).Return(cloud, nil).AnyTimes()
		})

		It("prints the deployment manifest and state file", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(stdOut).To(gbytes.Say("Deployment manifest: '/path/to/manifest.yml'"))
			Expect(stdOut).To(gbytes.Say("Deployment state: '/path/to/manifest-state.json'"))
		})

		It("does not migrate the legacy bosh-deployments.yml if manifest-state.json exists", func() {
			err := fakeFs.WriteFileString(deploymentStatePath, "{}")
			Expect(err).ToNot(HaveOccurred())

			expectLegacyMigrate.Times(0)

			err = command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("migrates the legacy bosh-deployments.yml if manifest-state.json does not exist", func() {
			err := fakeFs.RemoveAll(deploymentStatePath)
			Expect(err).ToNot(HaveOccurred())

			expectLegacyMigrate.Return(true, nil).Times(1)

			err = command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationParser.ParsePath).To(Equal(deploymentManifestPath))

			Expect(stdOut).To(gbytes.Say("Deployment manifest: '/path/to/manifest.yml'"))
			Expect(stdOut).To(gbytes.Say("Deployment state: '/path/to/manifest-state.json'"))
			Expect(stdOut).To(gbytes.Say("Migrated legacy deployments file: '/path/to/bosh-deployments.yml'"))
		})

		It("parses the installation manifest", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("parses the deployment manifest", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("validates release set manifest", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleaseSetValidator.ValidateInputs).To(Equal([]fakebirelsetmanifest.ValidateInput{
				{Manifest: releaseSetManifest},
			}))
		})

		It("validates installation manifest", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationValidator.ValidateInputs).To(Equal([]fakebiinstallmanifest.ValidateInput{
				{InstallationManifest: installationManifest, ReleaseSetManifest: releaseSetManifest},
			}))
		})

		It("validates bosh deployment manifest", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentValidator.ValidateInputs).To(Equal([]fakebideplval.ValidateInput{
				{Manifest: boshDeploymentManifest},
			}))
		})

		It("logs validating stages", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls[0]).To(Equal(fakebiui.PerformCall{
				Name: "validating",
				Stage: &fakebiui.FakeStage{
					PerformCalls: []fakebiui.PerformCall{
						{Name: "Validating releases"},
						{Name: "Validating deployment manifest"},
						{Name: "Validating stemcell"},
						{Name: "Validating cpi release"},
					},
				},
			}))
		})

		It("extracts CPI release tarball", func() {
			expectCPIReleaseExtract.Times(1)

			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		It("installs the CPI locally", func() {
			expectInstall.Times(1)
			expectNewCloud.Times(1)

			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		It("adds a new 'installing CPI' event logger stage", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls[1]).To(Equal(fakebiui.PerformCall{
				Name:  "installing CPI",
				Stage: &fakebiui.FakeStage{}, // mock installer doesn't add sub-stages
			}))
		})

		It("adds a new 'Starting registry' event logger stage", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls[2]).To(Equal(fakebiui.PerformCall{
				Name: "Starting registry",
			}))
		})

		Context("when the registry is configured", func() {
			BeforeEach(func() {
				installationManifest.Registry = biinstallmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
			})

			It("starts & stops the registry", func() {
				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)
				mockRegistryServer.EXPECT().Stop()

				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("deletes the extracted CPI release", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCPIRelease.DeleteCalled).To(BeTrue())
		})

		It("extracts the stemcell", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeStemcellExtractor.ExtractInputs).To(Equal([]fakebistemcell.ExtractInput{
				{TarballPath: stemcellTarballPath},
			}))
		})

		It("uploads the stemcell", func() {
			expectStemcellUpload.Times(1)

			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).ToNot(HaveOccurred())
		})

		It("adds a new 'deploying' event logger stage", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls[3]).To(Equal(fakebiui.PerformCall{
				Name:  "deploying",
				Stage: &fakebiui.FakeStage{}, // mock deployer doesn't add sub-stages
			}))
		})

		It("deploys", func() {
			expectDeploy.Times(1)

			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates the deployment record", func() {
			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())

			deploymentState, err := setupDeploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())

			Expect(deploymentState.CurrentManifestSHA1).To(Equal(manifestSHA1))
			Expect(deploymentState.Releases).To(Equal([]biconfig.ReleaseRecord{
				{
					ID:      "fake-uuid-0",
					Name:    fakeCPIRelease.Name(),
					Version: fakeCPIRelease.Version(),
				},
			}))
		})

		It("deletes unused stemcells", func() {
			expectStemcellDeleteUnused.Times(1)

			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when deployment has not changed", func() {
			JustBeforeEach(func() {
				previousDeploymentState := biconfig.DeploymentState{
					DirectorID:        directorID,
					CurrentReleaseIDs: []string{"my-release-id-1"},
					Releases: []biconfig.ReleaseRecord{{
						ID:      "my-release-id-1",
						Name:    fakeCPIRelease.Name(),
						Version: fakeCPIRelease.Version(),
					}},
					CurrentStemcellID: "my-stemcellRecordID",
					Stemcells: []biconfig.StemcellRecord{{
						ID:      "my-stemcellRecordID",
						Name:    cloudStemcell.Name(),
						Version: cloudStemcell.Version(),
					}},
					CurrentManifestSHA1: manifestSHA1,
				}

				err := setupDeploymentStateService.Save(previousDeploymentState)
				Expect(err).ToNot(HaveOccurred())
			})

			It("skips deploy", func() {
				expectDeploy.Times(0)

				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdOut).To(gbytes.Say("No deployment, stemcell or release changes. Skipping deploy."))
			})
		})

		Context("when parsing the cpi deployment manifest fails", func() {
			BeforeEach(func() {
				fakeDeploymentParser.ParseErr = bosherr.Error("fake-parse-error")
			})

			It("returns error", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing deployment manifest"))
				Expect(err.Error()).To(ContainSubstring("fake-parse-error"))
				Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
			})
		})

		Context("when the cpi release does not contain a 'cpi' job", func() {
			BeforeEach(func() {
				fakeCPIRelease.ReleaseJobs = []bireljob.Job{
					{
						Name: "not-cpi",
					},
				}
			})

			It("returns error", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid CPI release 'fake-cpi-release-name': CPI release must contain specified job 'fake-cpi-release-job-name'"))
			})
		})

		Context("when multiple releases are given", func() {
			var (
				otherReleaseTarballPath   string
				fakeOtherRelease          *fakebirel.FakeRelease
				expectOtherReleaseExtract *gomock.Call
			)

			BeforeEach(func() {
				otherReleaseTarballPath = "/path/to/other-release.tgz"

				fakeFs.WriteFileString(otherReleaseTarballPath, "")

				fakeOtherRelease = fakebirel.New("other-release", "1234")
				fakeOtherRelease.ReleaseJobs = []bireljob.Job{{Name: "not-cpi"}}

				deploymentReleases = []birel.Release{fakeOtherRelease, fakeCPIRelease}

				expectOtherReleaseExtract = mockReleaseExtractor.EXPECT().Extract(
					otherReleaseTarballPath,
				).Return(fakeOtherRelease, nil).AnyTimes()

				releaseSetManifest = birelsetmanifest.Manifest{
					Releases: []birelmanifest.ReleaseRef{
						{
							Name: "fake-cpi-release-name",
							URL:  "file://" + cpiReleaseTarballPath,
						},
						{
							Name: "other-release",
							URL:  "file://" + otherReleaseTarballPath,
						},
					},
				}
			})

			It("extracts all the release tarballs", func() {
				expectCPIReleaseExtract.Times(1)
				expectOtherReleaseExtract.Times(1)

				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			It("installs the CPI release locally", func() {
				expectInstall.Times(1)
				expectNewCloud.Times(1)

				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the deployment record", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).NotTo(HaveOccurred())

				deploymentState, err := setupDeploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentState.CurrentManifestSHA1).To(Equal(manifestSHA1))
				Expect(deploymentState.Releases).To(Equal([]biconfig.ReleaseRecord{
					{
						ID:      "fake-uuid-0",
						Name:    fakeCPIRelease.Name(),
						Version: fakeCPIRelease.Version(),
					},
					{
						ID:      "fake-uuid-1",
						Name:    fakeOtherRelease.Name(),
						Version: fakeOtherRelease.Version(),
					},
				}))
			})

			Context("when one of the releases in the deployment has changed", func() {
				JustBeforeEach(func() {
					olderReleaseVersion := "1233"
					Expect(fakeOtherRelease.Version()).ToNot(Equal(olderReleaseVersion))
					previousDeploymentState := biconfig.DeploymentState{
						DirectorID:        directorID,
						CurrentReleaseIDs: []string{"existing-release-id-1", "existing-release-id-2"},
						Releases: []biconfig.ReleaseRecord{
							{
								ID:      "existing-release-id-1",
								Name:    fakeCPIRelease.Name(),
								Version: fakeCPIRelease.Version(),
							},
							{
								ID:      "existing-release-id-2",
								Name:    fakeOtherRelease.Name(),
								Version: olderReleaseVersion,
							},
						},
						CurrentStemcellID: "my-stemcellRecordID",
						Stemcells: []biconfig.StemcellRecord{{
							ID:      "my-stemcellRecordID",
							Name:    cloudStemcell.Name(),
							Version: cloudStemcell.Version(),
						}},
						CurrentManifestSHA1: manifestSHA1,
					}

					err := setupDeploymentStateService.Save(previousDeploymentState)
					Expect(err).ToNot(HaveOccurred())
				})

				It("updates the deployment record, clearing out unused releases", func() {
					err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
					Expect(err).NotTo(HaveOccurred())

					deploymentState, err := setupDeploymentStateService.Load()
					Expect(err).ToNot(HaveOccurred())

					Expect(deploymentState.CurrentManifestSHA1).To(Equal(manifestSHA1))
					keys := []string{}
					ids := []string{}
					for _, releaseRecord := range deploymentState.Releases {
						keys = append(keys, fmt.Sprintf("%s-%s", releaseRecord.Name, releaseRecord.Version))
						ids = append(ids, releaseRecord.ID)
					}
					Expect(deploymentState.CurrentReleaseIDs).To(ConsistOf(ids))
					Expect(keys).To(ConsistOf([]string{
						fmt.Sprintf("%s-%s", fakeCPIRelease.Name(), fakeCPIRelease.Version()),
						fmt.Sprintf("%s-%s", fakeOtherRelease.Name(), fakeOtherRelease.Version()),
					}))
				})
			})

			Context("when the deployment has not changed", func() {
				JustBeforeEach(func() {
					previousDeploymentState := biconfig.DeploymentState{
						DirectorID:        directorID,
						CurrentReleaseIDs: []string{"my-release-id-1", "my-release-id-2"},
						Releases: []biconfig.ReleaseRecord{
							{
								ID:      "my-release-id-1",
								Name:    fakeCPIRelease.Name(),
								Version: fakeCPIRelease.Version(),
							},
							{
								ID:      "my-release-id-2",
								Name:    fakeOtherRelease.Name(),
								Version: fakeOtherRelease.Version(),
							},
						},
						CurrentStemcellID: "my-stemcellRecordID",
						Stemcells: []biconfig.StemcellRecord{{
							ID:      "my-stemcellRecordID",
							Name:    cloudStemcell.Name(),
							Version: cloudStemcell.Version(),
						}},
						CurrentManifestSHA1: manifestSHA1,
					}

					err := setupDeploymentStateService.Save(previousDeploymentState)
					Expect(err).ToNot(HaveOccurred())
				})

				It("skips deploy", func() {
					expectDeploy.Times(0)

					err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
					Expect(err).NotTo(HaveOccurred())
					Expect(stdOut).To(gbytes.Say("No deployment, stemcell or release changes. Skipping deploy."))
				})
			})
		})

		Context("when release name does not match the name in release tarball", func() {
			BeforeEach(func() {
				releaseSetManifest.Releases = []birelmanifest.ReleaseRef{
					{
						Name: "fake-other-cpi-release-name",
						URL:  "file://" + cpiReleaseTarballPath,
					},
				}
			})

			It("returns an error", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Release name 'fake-other-cpi-release-name' does not match the name in release tarball 'fake-cpi-release-name'"))
			})
		})

		Context("When the stemcell tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(stemcellTarballPath)
			})

			It("returns error", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the stemcell '/stemcell/tarball/path' exists"))

				performCall := fakeStage.PerformCalls[0].Stage.PerformCalls[2]
				Expect(performCall.Name).To(Equal("Validating stemcell"))
				Expect(performCall.Error.Error()).To(Equal("Verifying that the stemcell '/stemcell/tarball/path' exists"))
			})
		})

		Context("When the CPI release tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(cpiReleaseTarballPath)
			})

			It("returns error", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the release '/release/tarball/path' exists"))

				performCall := fakeStage.PerformCalls[0].Stage.PerformCalls[0]
				Expect(performCall.Name).To(Equal("Validating releases"))
				Expect(performCall.Error.Error()).To(Equal("Verifying that the release '/release/tarball/path' exists"))
			})
		})

		Context("when the deployment state file does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(deploymentStatePath)
			})

			It("creates a deployment state", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).ToNot(HaveOccurred())

				deploymentState, err := setupDeploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentState.DirectorID).To(Equal(directorID))
			})
		})

		It("returns err when the deployment manifest does not exist", func() {
			fakeFs.RemoveAll(deploymentManifestPath)

			err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Deployment manifest does not exist at '/path/to/manifest.yml'"))
			Expect(stdErr).To(gbytes.Say("Deployment '/path/to/manifest.yml' does not exist"))
		})

		Context("when the installation manifest is invalid", func() {
			BeforeEach(func() {
				fakeInstallationValidator.SetValidateBehavior([]fakebiinstallmanifest.ValidateOutput{
					{Err: bosherr.Error("fake-installation-validation-error")},
				})
			})

			It("returns err", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-installation-validation-error"))
			})

			It("logs the failed event log", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())

				performCall := fakeStage.PerformCalls[0].Stage.PerformCalls[1]
				Expect(performCall.Name).To(Equal("Validating deployment manifest"))
				Expect(performCall.Error.Error()).To(Equal("Validating installation manifest: fake-installation-validation-error"))
			})
		})

		Context("when the deployment manifest is invalid", func() {
			BeforeEach(func() {
				fakeDeploymentValidator.SetValidateBehavior([]fakebideplval.ValidateOutput{
					{Err: bosherr.Error("fake-deployment-validation-error")},
				})
			})

			It("returns err", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-deployment-validation-error"))
			})

			It("logs the failed event log", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())

				performCall := fakeStage.PerformCalls[0].Stage.PerformCalls[1]
				Expect(performCall.Name).To(Equal("Validating deployment manifest"))
				Expect(performCall.Error.Error()).To(Equal("Validating deployment manifest: fake-deployment-validation-error"))
			})
		})

		It("returns err when number of arguments is not equal 2", func() {
			err := command.Run(fakeStage, []string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid usage"))

			err = command.Run(fakeStage, []string{"1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid usage"))

			err = command.Run(fakeStage, []string{"1", "2", "3"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid usage"))
		})

		Context("when uploading stemcell fails", func() {
			JustBeforeEach(func() {
				expectStemcellUpload.Return(nil, bosherr.Error("fake-upload-error"))
			})

			It("returns an error", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-upload-error"))
			})
		})

		Context("when deploy fails", func() {
			BeforeEach(func() {
				mockDeployer.EXPECT().Deploy(
					cloud,
					boshDeploymentManifest,
					cloudStemcell,
					installationManifest.Registry,
					installationManifest.SSHTunnel,
					fakeVMManager,
					mockBlobstore,
					gomock.Any(),
				).Return(nil, errors.New("fake-deploy-error")).AnyTimes()

				previousDeploymentState := biconfig.DeploymentState{
					CurrentReleaseIDs: []string{"my-release-id-1"},
					Releases: []biconfig.ReleaseRecord{{
						ID:      "my-release-id-1",
						Name:    fakeCPIRelease.Name(),
						Version: fakeCPIRelease.Version(),
					}},
					CurrentManifestSHA1: "fake-manifest-sha",
				}

				setupDeploymentStateService.Save(previousDeploymentState)
			})

			It("clears the deployment record", func() {
				err := command.Run(fakeStage, []string{deploymentManifestPath, stemcellTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-deploy-error"))

				deploymentState, err := setupDeploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentState.CurrentManifestSHA1).To(Equal(""))
				Expect(deploymentState.Releases).To(Equal([]biconfig.ReleaseRecord{}))
				Expect(deploymentState.CurrentReleaseIDs).To(Equal([]string{}))
			})
		})
	})
}
