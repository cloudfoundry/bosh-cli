package integration_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"os"

	"code.google.com/p/gomock/gomock"
	mock_cloud "github.com/cloudfoundry/bosh-micro-cli/cloud/mocks"
	mock_cpi "github.com/cloudfoundry/bosh-micro-cli/cpi/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/mocks"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmdeployer "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakebmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi/fakes"
	fakebmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("bosh-micro", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Deploy", func() {
		var (
			fs     *fakesys.FakeFileSystem
			logger boshlog.Logger

			mockCPIDeploymentFactory  *mock_cpi.MockDeploymentFactory
			mockRegistryServerManager *mock_registry.MockServerManager
			mockRegistryServer        *mock_registry.MockServer

			fakeCPIInstaller        *fakebmcpi.FakeInstaller
			fakeStemcellExtractor   *fakebmstemcell.FakeExtractor
			fakeUUIDGenerator       *fakeuuid.FakeGenerator
			fakeSHA1Calculator      *fakebmcrypto.FakeSha1Calculator
			deploymentConfigService bmconfig.DeploymentConfigService
			vmRepo                  bmconfig.VMRepo
			diskRepo                bmconfig.DiskRepo
			stemcellRepo            bmconfig.StemcellRepo
			deploymentRepo          bmconfig.DeploymentRepo
			releaseRepo             bmconfig.ReleaseRepo
			userConfig              bmconfig.UserConfig

			sshTunnelFactory bmsshtunnel.Factory

			diskManagerFactory bmdisk.ManagerFactory
			diskDeployer       bminstance.DiskDeployer

			ui          *fakeui.FakeUI
			eventLogger bmeventlog.EventLogger

			stemcellManagerFactory bmstemcell.ManagerFactory
			vmManagerFactory       bmvm.ManagerFactory

			fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
			fakeTemplatesSpecGenerator *fakebmas.FakeTemplatesSpecGenerator
			applySpec                  bmas.ApplySpec

			mockAgentClient        *mock_agentclient.MockAgentClient
			mockAgentClientFactory *mock_agentclient.MockFactory
			mockCloud              *mock_cloud.MockCloud
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			deploymentConfigPath   = "/fake-bosh-deployments.json"

			cloudProperties   = map[string]interface{}{}
			stemcellImagePath = "fake-stemcell-image-path"
			stemcellCID       = "fake-stemcell-cid"
			env               = map[string]interface{}{}
			networksSpec      = map[string]interface{}{
				"network-1": map[string]interface{}{
					"type":             "dynamic",
					"ip":               "",
					"cloud_properties": cloudProperties,
				},
			}
			agentRunningState = bmac.State{JobState: "running"}
			mbusURL           = "http://fake-mbus-url"
		)

		var writeDeploymentManifest = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---
name: test-release

networks:
- name: network-1
  type: dynamic

resource_pools:
- name: resource-pool-1
  network: network-1

jobs:
- name: cpi
  instances: 1
  persistent_disk: 1024
  networks:
  - name: network-1

cloud_provider:
  mbus: http://fake-mbus-url
`)
			Expect(err).ToNot(HaveOccurred())

			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
				deploymentManifestPath: {Sha1: "fake-deployment-sha1-1"},
			})
		}

		var writeDeploymentManifestWithLargerDisk = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---
name: test-release

networks:
- name: network-1
  type: dynamic

resource_pools:
- name: resource-pool-1
  network: network-1

jobs:
- name: cpi
  instances: 1
  persistent_disk: 2048
  networks:
  - name: network-1

cloud_provider:
  mbus: http://fake-mbus-url
`)
			Expect(err).ToNot(HaveOccurred())

			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
				deploymentManifestPath: {Sha1: "fake-deployment-sha1-2"},
			})
		}

		var writeCPIReleaseTarball = func() {
			err := fs.WriteFileString("/fake-cpi-release.tgz", "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowCPIToBeInstalled = func() {
			cpiRelease := bmrel.NewRelease(
				"fake-cpi-release-name",
				"fake-cpi-release-version",
				[]bmrel.Job{},
				[]*bmrel.Package{},
				"fake-cpi-extracted-dir",
				fs,
			)
			fakeCPIInstaller.SetExtractBehavior("/fake-cpi-release.tgz", func(releaseTarballPath string) (bmrel.Release, error) {
				err := fs.MkdirAll("fake-cpi-extracted-dir", os.ModePerm)
				return cpiRelease, err
			})

			cpiDeploymentManifest := bmdepl.CPIDeploymentManifest{
				Name: "test-release",
				Mbus: mbusURL,
			}
			fakeCPIInstaller.SetInstallBehavior(cpiDeploymentManifest, cpiRelease, mockCloud, nil)

			cpiDeployment := bmcpi.NewDeployment(cpiDeploymentManifest, mockRegistryServerManager, fakeCPIInstaller)
			mockCPIDeploymentFactory.EXPECT().NewDeployment(cpiDeploymentManifest).Return(cpiDeployment).AnyTimes()
		}

		var writeStemcellReleaseTarball = func() {
			err := fs.WriteFileString("/fake-stemcell-release.tgz", "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowStemcellToBeExtracted = func() {
			stemcellManifest := bmstemcell.Manifest{
				ImagePath: "fake-stemcell-image-path",
				Name:      "fake-stemcell-name",
				Version:   "fake-stemcell-version",
				SHA1:      "fake-stemcell-sha1",
			}
			stemcellApplySpec := bmstemcell.ApplySpec{
				Job: bmstemcell.Job{
					Name:      "cpi",
					Templates: []bmstemcell.Blob{},
				},
				Packages: map[string]bmstemcell.Blob{},
				Networks: map[string]interface{}{},
			}
			extractedStemcell := bmstemcell.NewExtractedStemcell(
				stemcellManifest,
				stemcellApplySpec,
				"fake-stemcell-extracted-dir",
				fs,
			)
			fakeStemcellExtractor.SetExtractBehavior("/fake-stemcell-release.tgz", extractedStemcell, nil)
		}

		var allowApplySpecToBeCreated = func() {
			applySpec = bmas.ApplySpec{
				Deployment: "",
				Index:      0,
				Packages:   map[string]bmas.Blob{},
				Networks:   map[string]interface{}{},
				Job:        bmas.Job{},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{},
				ConfigurationHash:        "",
			}
			fakeApplySpecFactory.CreateApplySpec = applySpec
		}

		var newDeployCmd = func() Cmd {
			deploymentParser := bmdepl.NewParser(fs, logger)

			boshDeploymentValidator := bmdeplval.NewBoshDeploymentValidator()

			deploymentRecord := bmdeployer.NewDeploymentRecord(deploymentRepo, releaseRepo, stemcellRepo, fakeSHA1Calculator)

			deployer := bmdeployer.NewDeployer(
				stemcellManagerFactory,
				vmManagerFactory,
				sshTunnelFactory,
				diskDeployer,
				eventLogger,
				logger,
			)

			deploymentFactory := bmdeployer.NewFactory(deployer)

			return NewDeployCmd(
				ui,
				userConfig,
				fs,
				deploymentParser,
				boshDeploymentValidator,
				mockCPIDeploymentFactory,
				fakeStemcellExtractor,
				deploymentRecord,
				deploymentFactory,
				eventLogger,
				logger,
			)
		}

		var expectDeployFlow = func() {
			vmCID := "fake-vm-cid-1"
			diskCID := "fake-disk-cid-1"
			diskSize := 1024

			gomock.InOrder(
				mockCloud.EXPECT().CreateStemcell(cloudProperties, stemcellImagePath).Return(stemcellCID, nil),
				mockCloud.EXPECT().CreateVM(stemcellCID, cloudProperties, networksSpec, env).Return(vmCID, nil),
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				mockCloud.EXPECT().CreateDisk(diskSize, cloudProperties, vmCID).Return(diskCID, nil),
				mockCloud.EXPECT().AttachDisk(vmCID, diskCID),
				mockAgentClient.EXPECT().MountDisk(diskCID),

				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		var expectDeployWithMigration = func() {
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			gomock.InOrder(
				// shutdown old vm
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(stemcellCID, cloudProperties, networksSpec, env).Return(newVMCID, nil),
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		var expectDeployWithMigrationFailure = func() {
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			gomock.InOrder(
				// shutdown old vm
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(stemcellCID, cloudProperties, networksSpec, env).Return(newVMCID, nil),
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate (with error
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk().Return(errors.New("fake-migration-error")),
			)
		}

		var expectDeployWithMigrationRepair = func() {
			oldVMCID := "fake-vm-cid-2"
			newVMCID := "fake-vm-cid-3"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-3"
			newDiskSize := 2048

			gomock.InOrder(
				// shutdown old vm
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(stemcellCID, cloudProperties, networksSpec, env).Return(newVMCID, nil),
				mockAgentClientFactory.EXPECT().Create(mbusURL).Return(mockAgentClient),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			logger = boshlog.NewLogger(boshlog.LevelNone)
			deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, fs, logger)
			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()

			fakeSHA1Calculator = fakebmcrypto.NewFakeSha1Calculator()

			mockCPIDeploymentFactory = mock_cpi.NewMockDeploymentFactory(mockCtrl)

			sshTunnelFactory = bmsshtunnel.NewFactory(logger)

			vmRepo = bmconfig.NewVMRepo(deploymentConfigService)
			diskRepo = bmconfig.NewDiskRepo(deploymentConfigService, fakeUUIDGenerator)
			stemcellRepo = bmconfig.NewStemcellRepo(deploymentConfigService, fakeUUIDGenerator)
			deploymentRepo = bmconfig.NewDeploymentRepo(deploymentConfigService)
			releaseRepo = bmconfig.NewReleaseRepo(deploymentConfigService, fakeUUIDGenerator)

			diskManagerFactory = bmdisk.NewManagerFactory(diskRepo, logger)
			diskDeployer = bminstance.NewDiskDeployer(diskManagerFactory, diskRepo, logger)

			mockCloud = mock_cloud.NewMockCloud(mockCtrl)

			mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)
			mockRegistryServer = mock_registry.NewMockServer(mockCtrl)

			fakeCPIInstaller = fakebmcpi.NewFakeInstaller()
			fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()

			ui = &fakeui.FakeUI{}
			eventLogger = bmeventlog.NewEventLogger(ui)

			mockAgentClientFactory = mock_agentclient.NewMockFactory(mockCtrl)
			mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)

			stemcellManagerFactory = bmstemcell.NewManagerFactory(stemcellRepo, eventLogger)

			fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()
			fakeTemplatesSpecGenerator = fakebmas.NewFakeTemplatesSpecGenerator()

			vmManagerFactory = bmvm.NewManagerFactory(
				vmRepo,
				stemcellRepo,
				mockAgentClientFactory,
				fakeApplySpecFactory,
				fakeTemplatesSpecGenerator,
				fs,
				logger,
			)

			userConfig = bmconfig.UserConfig{DeploymentFile: deploymentManifestPath}

			writeDeploymentManifest()
			writeCPIReleaseTarball()
			allowCPIToBeInstalled()

			writeStemcellReleaseTarball()
			allowStemcellToBeExtracted()
			allowApplySpecToBeCreated()
		})

		Context("when the deployment has not been set", func() {
			BeforeEach(func() {
				userConfig.DeploymentFile = ""
			})

			It("returns an error", func() {
				err := newDeployCmd().Run([]string{"/fake-cpi-release.tgz", "/fake-stemcell-release.tgz"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No deployment set"))
			})
		})

		Context("when the deployment config file does not exist", func() {
			BeforeEach(func() {
				err := fs.RemoveAll(deploymentConfigPath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("creates one", func() {
				expectDeployFlow()

				err := newDeployCmd().Run([]string{"/fake-cpi-release.tgz", "/fake-stemcell-release.tgz"})
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists(deploymentConfigPath)).To(BeTrue())
			})
		})

		Context("when the deployment has been deployed", func() {
			BeforeEach(func() {
				expectDeployFlow()

				err := newDeployCmd().Run([]string{"/fake-cpi-release.tgz", "/fake-stemcell-release.tgz"})
				Expect(err).ToNot(HaveOccurred())

				// reset output buffer
				ui.Said = []string{}
			})

			Context("when persistent disk size is increased", func() {
				BeforeEach(func() {
					writeDeploymentManifestWithLargerDisk()
				})

				It("migrates the disk content", func() {
					expectDeployWithMigration()

					err := newDeployCmd().Run([]string{"/fake-cpi-release.tgz", "/fake-stemcell-release.tgz"})
					Expect(err).ToNot(HaveOccurred())
				})

				Context("after migration has failed", func() {
					BeforeEach(func() {
						expectDeployWithMigrationFailure()

						err := newDeployCmd().Run([]string{"/fake-cpi-release.tgz", "/fake-stemcell-release.tgz"})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migration-error"))

						// reset output buffer
						ui.Said = []string{}

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(HaveLen(2)) // current + unused
					})

					It("deletes unused disks", func() {
						expectDeployWithMigrationRepair()

						mockCloud.EXPECT().DeleteDisk("fake-disk-cid-2")

						err := newDeployCmd().Run([]string{"/fake-cpi-release.tgz", "/fake-stemcell-release.tgz"})
						Expect(err).ToNot(HaveOccurred())

						diskRecord, found, err := diskRepo.FindCurrent()
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(BeTrue())
						Expect(diskRecord.CID).To(Equal("fake-disk-cid-3"))

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(Equal([]bmconfig.DiskRecord{diskRecord}))
					})
				})
			})
		})
	})
})
