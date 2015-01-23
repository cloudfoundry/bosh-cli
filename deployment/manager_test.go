package deployment_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_cloud "github.com/cloudfoundry/bosh-micro-cli/cloud/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"
	mock_disk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk/mocks"
	mock_instance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/mocks"
	mock_deployment "github.com/cloudfoundry/bosh-micro-cli/deployment/mocks"
	mock_stemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakebmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Manager", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("FindCurrent", func() {
		var (
			mockInstanceManager   *mock_instance.MockManager
			mockDiskManager       *mock_disk.MockManager
			mockStemcellManager   *mock_stemcell.MockManager
			mockDeploymentFactory *mock_deployment.MockFactory
			mockDeployment        *mock_deployment.MockDeployment

			deploymentManager Manager

			expectedInstances []bminstance.Instance
			expectedDisks     []bmdisk.Disk
			expectedStemcells []bmstemcell.CloudStemcell

			expectNewDeployment *gomock.Call
		)

		BeforeEach(func() {
			mockInstanceManager = mock_instance.NewMockManager(mockCtrl)
			mockDiskManager = mock_disk.NewMockManager(mockCtrl)
			mockStemcellManager = mock_stemcell.NewMockManager(mockCtrl)
			mockDeploymentFactory = mock_deployment.NewMockFactory(mockCtrl)
			mockDeployment = mock_deployment.NewMockDeployment(mockCtrl)

			expectedInstances = []bminstance.Instance{}
			expectedDisks = []bmdisk.Disk{}
			expectedStemcells = []bmstemcell.CloudStemcell{}
		})

		JustBeforeEach(func() {
			mockInstanceManager.EXPECT().FindCurrent().Return(expectedInstances, nil)
			mockDiskManager.EXPECT().FindCurrent().Return(expectedDisks, nil)
			mockStemcellManager.EXPECT().FindCurrent().Return(expectedStemcells, nil)

			expectNewDeployment = mockDeploymentFactory.EXPECT().NewDeployment(expectedInstances, expectedDisks, expectedStemcells).Return(mockDeployment).AnyTimes()

			deploymentManager = NewManager(mockInstanceManager, mockDiskManager, mockStemcellManager, mockDeploymentFactory)
		})

		Context("when no current instances, disks, or stemcells exist", func() {
			It("returns not found", func() {
				_, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when current instances exist", func() {
			BeforeEach(func() {
				instance := mock_instance.NewMockInstance(mockCtrl)
				expectedInstances = append(expectedInstances, instance)
			})

			It("returns a deployment that wraps the current instances, disks, & stemcells", func() {
				expectNewDeployment.Times(1)

				deployment, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(deployment).To(Equal(mockDeployment))
			})
		})

		Context("when current disk exist", func() {
			BeforeEach(func() {
				disk := mock_disk.NewMockDisk(mockCtrl)
				expectedDisks = append(expectedDisks, disk)
			})

			It("returns a deployment that wraps the current instances, disks, & stemcells", func() {
				expectNewDeployment.Times(1)

				deployment, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(deployment).To(Equal(mockDeployment))
			})
		})

		Context("when current stemcell exist", func() {
			BeforeEach(func() {
				stemcell := mock_stemcell.NewMockCloudStemcell(mockCtrl)
				expectedStemcells = append(expectedStemcells, stemcell)
			})

			It("returns a deployment that wraps the current instances, disks, & stemcells", func() {
				expectNewDeployment.Times(1)

				deployment, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(deployment).To(Equal(mockDeployment))
			})
		})
	})

	Describe("Cleanup", func() {
		var (
			logger boshlog.Logger
			fs     boshsys.FileSystem

			mockDeploymentFactory *mock_deployment.MockFactory

			mockStateBuilderFactory *mock_instance.MockStateBuilderFactory

			mockBlobstore *mock_blobstore.MockBlobstore

			fakeUUIDGenerator       *fakeuuid.FakeGenerator
			fakeRepoUUIDGenerator   *fakeuuid.FakeGenerator
			deploymentConfigService bmconfig.DeploymentConfigService
			vmRepo                  bmconfig.VMRepo
			diskRepo                bmconfig.DiskRepo
			stemcellRepo            bmconfig.StemcellRepo

			mockCloud       *mock_cloud.MockCloud
			mockAgentClient *mock_agentclient.MockAgentClient

			deploymentConfigPath = "/deployment.json"

			fakeStage *fakebmeventlog.FakeStage

			deploymentManager Manager
		)

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fs = fakesys.NewFakeFileSystem()

			mockDeploymentFactory = mock_deployment.NewMockFactory(mockCtrl)

			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, fs, fakeUUIDGenerator, logger)

			fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()
			vmRepo = bmconfig.NewVMRepo(deploymentConfigService)
			diskRepo = bmconfig.NewDiskRepo(deploymentConfigService, fakeRepoUUIDGenerator)
			stemcellRepo = bmconfig.NewStemcellRepo(deploymentConfigService, fakeRepoUUIDGenerator)

			mockCloud = mock_cloud.NewMockCloud(mockCtrl)
			mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)

			fakeStage = fakebmeventlog.NewFakeStage()
		})

		JustBeforeEach(func() {
			diskManagerFactory := bmdisk.NewManagerFactory(diskRepo, logger)
			diskDeployer := bmvm.NewDiskDeployer(diskManagerFactory, diskRepo, logger)

			vmManagerFactory := bmvm.NewManagerFactory(vmRepo, stemcellRepo, diskDeployer, fakeUUIDGenerator, fs, logger)
			sshTunnelFactory := bmsshtunnel.NewFactory(logger)

			mockStateBuilderFactory = mock_instance.NewMockStateBuilderFactory(mockCtrl)

			instanceFactory := bminstance.NewFactory(mockStateBuilderFactory)
			instanceManagerFactory := bminstance.NewManagerFactory(sshTunnelFactory, instanceFactory, logger)
			stemcellManagerFactory := bmstemcell.NewManagerFactory(stemcellRepo)

			mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)

			deploymentManagerFactory := NewManagerFactory(vmManagerFactory, instanceManagerFactory, diskManagerFactory, stemcellManagerFactory, mockDeploymentFactory)
			deploymentManager = deploymentManagerFactory.NewManager(mockCloud, mockAgentClient, mockBlobstore)
		})

		Context("no orphan disk or stemcell records exist", func() {
			var (
				currentDiskRecord     bmconfig.DiskRecord
				currentStemcellRecord bmconfig.StemcellRecord
			)

			BeforeEach(func() {
				var err error
				currentDiskRecord, err = diskRepo.Save("fake-disk-cid", 100, nil)
				Expect(err).ToNot(HaveOccurred())
				err = diskRepo.UpdateCurrent(currentDiskRecord.ID)
				Expect(err).ToNot(HaveOccurred())

				currentStemcellRecord, err = stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid")
				Expect(err).ToNot(HaveOccurred())
				err = stemcellRepo.UpdateCurrent(currentStemcellRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not delete anything", func() {
				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				diskRecord, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(diskRecord).To(Equal(currentDiskRecord))

				stemcellRecord, found, err := stemcellRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(stemcellRecord).To(Equal(currentStemcellRecord))
			})

			It("does not log any stages", func() {
				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(BeEmpty())
			})
		})

		Context("orphan disk records exist", func() {
			BeforeEach(func() {
				_, err := diskRepo.Save("orphan-disk-cid", 100, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the unused disks", func() {
				mockCloud.EXPECT().DeleteDisk("orphan-disk-cid")

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				diskRecords, err := diskRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(diskRecords).To(BeEmpty(), "expected no disk records")
			})

			It("logs delete stage", func() {
				mockCloud.EXPECT().DeleteDisk("orphan-disk-cid")

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(Equal([]*fakebmeventlog.FakeStep{
					&fakebmeventlog.FakeStep{
						Name: "Deleting unused disk 'orphan-disk-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
				}))
			})

			Context("when disks have been deleted manually (in the infrastructure)", func() {
				It("deletes the unused disks, ignoring DiskNotFoundError", func() {
					mockCloud.EXPECT().DeleteDisk("orphan-disk-cid").Return(bmcloud.NewCPIError("delete_disk", bmcloud.CmdError{
						Type:    bmcloud.DiskNotFoundError,
						Message: "fake-disk-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					diskRecords, err := diskRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(diskRecords).To(BeEmpty(), "expected no disk records")
				})

				It("logs disk deletion as skipped", func() {
					mockCloud.EXPECT().DeleteDisk("orphan-disk-cid").Return(bmcloud.NewCPIError("delete_disk", bmcloud.CmdError{
						Type:    bmcloud.DiskNotFoundError,
						Message: "fake-disk-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.Steps).To(Equal([]*fakebmeventlog.FakeStep{
						&fakebmeventlog.FakeStep{
							Name: "Deleting unused disk 'orphan-disk-cid'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Skipped,
							},
							SkipMessage: "CPI 'delete_disk' method responded with error: CmdError{\"type\":\"Bosh::Cloud::DiskNotFound\",\"message\":\"fake-disk-not-found-message\",\"ok_to_retry\":false}",
						},
					}))
				})
			})
		})

		Context("orphan stemcell records exist", func() {
			BeforeEach(func() {
				_, err := stemcellRepo.Save("orphan-stemcell-name", "orphan-stemcell-version", "orphan-stemcell-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the unused stemcells", func() {
				mockCloud.EXPECT().DeleteStemcell("orphan-stemcell-cid")

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				stemcellRecords, err := stemcellRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
			})

			It("logs delete stage", func() {
				mockCloud.EXPECT().DeleteStemcell("orphan-stemcell-cid")

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(Equal([]*fakebmeventlog.FakeStep{
					&fakebmeventlog.FakeStep{
						Name: "Deleting unused stemcell 'orphan-stemcell-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
				}))
			})

			Context("when stemcells have been deleted manually (in the infrastructure)", func() {
				It("deletes the unused stemcells, ignoring StemcellNotFoundError", func() {
					mockCloud.EXPECT().DeleteStemcell("orphan-stemcell-cid").Return(bmcloud.NewCPIError("delete_stemcell", bmcloud.CmdError{
						Type:    bmcloud.StemcellNotFoundError,
						Message: "fake-stemcell-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					stemcellRecords, err := diskRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
				})

				It("logs stemcell deletion as skipped", func() {
					mockCloud.EXPECT().DeleteStemcell("orphan-stemcell-cid").Return(bmcloud.NewCPIError("delete_stemcell", bmcloud.CmdError{
						Type:    bmcloud.StemcellNotFoundError,
						Message: "fake-stemcell-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.Steps).To(Equal([]*fakebmeventlog.FakeStep{
						&fakebmeventlog.FakeStep{
							Name: "Deleting unused stemcell 'orphan-stemcell-cid'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Skipped,
							},
							SkipMessage: "CPI 'delete_stemcell' method responded with error: CmdError{\"type\":\"Bosh::Cloud::StemcellNotFound\",\"message\":\"fake-stemcell-not-found-message\",\"ok_to_retry\":false}",
						},
					}))
				})
			})
		})
	})
})
