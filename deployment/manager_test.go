package deployment_test

import (
	"github.com/cloudfoundry/bosh-cli/v7/stemcell/stemcellfakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/agentclient/agentclientfakes"
	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/deploymentfakes"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/disk/diskfakes"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/instance/instancefakes"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state/statefakes"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("Manager", func() {
	Describe("FindCurrent", func() {
		var (
			mockInstanceManager   *instancefakes.FakeManager
			mockDiskManager       *diskfakes.FakeManager
			mockStemcellManager   *stemcellfakes.FakeManager
			mockDeploymentFactory *deploymentfakes.FakeFactory
			mockDeployment        *deploymentfakes.FakeDeployment

			deploymentManager Manager

			expectedInstances []biinstance.Instance
			expectedDisks     []bidisk.Disk
			expectedStemcells []bistemcell.CloudStemcell
		)

		BeforeEach(func() {
			mockInstanceManager = &instancefakes.FakeManager{}
			mockDiskManager = &diskfakes.FakeManager{}
			mockStemcellManager = &stemcellfakes.FakeManager{}
			mockDeploymentFactory = &deploymentfakes.FakeFactory{}
			mockDeployment = &deploymentfakes.FakeDeployment{}

			expectedInstances = []biinstance.Instance{}
			expectedDisks = []bidisk.Disk{}
			expectedStemcells = []bistemcell.CloudStemcell{}
		})

		JustBeforeEach(func() {
			mockInstanceManager.FindCurrentReturns(expectedInstances, nil)
			mockDiskManager.FindCurrentReturns(expectedDisks, nil)
			mockStemcellManager.FindCurrentReturns(expectedStemcells, nil)

			mockDeploymentFactory.NewDeploymentReturns(mockDeployment)

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
				instance := &instancefakes.FakeInstance{}
				expectedInstances = append(expectedInstances, instance)
			})

			It("returns a deployment that wraps the current instances, disks, & stemcells", func() {
				deployment, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(deployment).To(Equal(mockDeployment))

				Expect(mockDeploymentFactory.NewDeploymentCallCount()).To(Equal(1))
				instances, disks, stemcells := mockDeploymentFactory.NewDeploymentArgsForCall(0)
				Expect(instances).To(Equal(expectedInstances))
				Expect(disks).To(Equal(expectedDisks))
				Expect(stemcells).To(Equal(expectedStemcells))
			})
		})

		Context("when current disk exist", func() {
			BeforeEach(func() {
				disk := &diskfakes.FakeDisk{}
				expectedDisks = append(expectedDisks, disk)
			})

			It("returns a deployment that wraps the current instances, disks, & stemcells", func() {
				deployment, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(deployment).To(Equal(mockDeployment))

				Expect(mockDeploymentFactory.NewDeploymentCallCount()).To(Equal(1))
			})
		})

		Context("when current stemcell exist", func() {
			BeforeEach(func() {
				stemcell := stemcellfakes.NewFakeCloudStemcell("fake-stemcell-cid", "fake-stemcell-name", "fake-stemcell-version", 1)
				expectedStemcells = append(expectedStemcells, stemcell)
			})

			It("returns a deployment that wraps the current instances, disks, & stemcells", func() {
				deployment, found, err := deploymentManager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(deployment).To(Equal(mockDeployment))

				Expect(mockDeploymentFactory.NewDeploymentCallCount()).To(Equal(1))
			})
		})
	})

	Describe("Cleanup", func() {
		var (
			logger boshlog.Logger
			fs     boshsys.FileSystem

			mockDeploymentFactory *deploymentfakes.FakeFactory

			mockStateBuilderFactory *statefakes.FakeBuilderFactory

			mockBlobstore *blobstorefakes.FakeBlobstore

			fakeUUIDGenerator      *fakeuuid.FakeGenerator
			fakeRepoUUIDGenerator  *fakeuuid.FakeGenerator
			deploymentStateService biconfig.DeploymentStateService
			vmRepo                 biconfig.VMRepo
			diskRepo               biconfig.DiskRepo
			stemcellRepo           biconfig.StemcellRepo

			mockCloud       *cloudfakes.FakeCloud
			mockAgentClient *agentclientfakes.FakeAgentClient

			fakeStage *testui.Stage

			deploymentManager  Manager
			stemcellApiVersion = 2
		)

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fs = fakesys.NewFakeFileSystem()

			mockDeploymentFactory = &deploymentfakes.FakeFactory{}

			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			deploymentStateService = biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, "/deployment.json")

			fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()
			vmRepo = biconfig.NewVMRepo(deploymentStateService)
			diskRepo = biconfig.NewDiskRepo(deploymentStateService, fakeRepoUUIDGenerator)
			stemcellRepo = biconfig.NewStemcellRepo(deploymentStateService, fakeRepoUUIDGenerator)

			mockCloud = &cloudfakes.FakeCloud{}
			mockAgentClient = &agentclientfakes.FakeAgentClient{}

			fakeStage = &testui.Stage{}
		})

		JustBeforeEach(func() {
			diskManagerFactory := bidisk.NewManagerFactory(diskRepo, logger)
			diskDeployer := bivm.NewDiskDeployer(diskManagerFactory, diskRepo, logger, false)

			vmManagerFactory := bivm.NewManagerFactory(vmRepo, stemcellRepo, diskDeployer, fakeUUIDGenerator, fs, logger)
			sshTunnelFactory := bisshtunnel.NewFactory(logger)

			mockStateBuilderFactory = &statefakes.FakeBuilderFactory{}

			instanceFactory := biinstance.NewFactory(mockStateBuilderFactory)
			instanceManagerFactory := biinstance.NewManagerFactory(sshTunnelFactory, instanceFactory, logger)
			stemcellManagerFactory := bistemcell.NewManagerFactory(stemcellRepo)

			mockBlobstore = &blobstorefakes.FakeBlobstore{}

			deploymentManagerFactory := NewManagerFactory(vmManagerFactory, instanceManagerFactory, diskManagerFactory, stemcellManagerFactory, mockDeploymentFactory)
			deploymentManager = deploymentManagerFactory.NewManager(mockCloud, mockAgentClient, mockBlobstore)
		})

		Context("no orphan disk or stemcell records exist", func() {
			var (
				currentDiskRecord     biconfig.DiskRecord
				currentStemcellRecord biconfig.StemcellRecord
			)

			BeforeEach(func() {
				var err error
				currentDiskRecord, err = diskRepo.Save("fake-disk-cid", 100, nil)
				Expect(err).ToNot(HaveOccurred())
				err = diskRepo.UpdateCurrent(currentDiskRecord.ID)
				Expect(err).ToNot(HaveOccurred())

				currentStemcellRecord, err = stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid", stemcellApiVersion)
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

				Expect(fakeStage.PerformCalls).To(BeEmpty())
			})
		})

		Context("orphan disk records exist", func() {
			BeforeEach(func() {
				_, err := diskRepo.Save("orphan-disk-cid", 100, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the unused disks", func() {
				mockCloud.DeleteDiskReturns(nil)

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				diskRecords, err := diskRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(diskRecords).To(BeEmpty(), "expected no disk records")
			})

			It("logs delete stage", func() {
				mockCloud.DeleteDiskReturns(nil)

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(ContainElement(&testui.PerformCall{
					Name: "Deleting unused disk 'orphan-disk-cid'",
				}))
			})

			Context("when disks have been deleted manually (in the infrastructure)", func() {
				It("deletes the unused disks, ignoring DiskNotFoundError", func() {
					mockCloud.DeleteDiskReturns(bicloud.NewCPIError("delete_disk", bicloud.CmdError{
						Type:    bicloud.DiskNotFoundError,
						Message: "fake-disk-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					diskRecords, err := diskRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(diskRecords).To(BeEmpty(), "expected no disk records")
				})

				It("logs disk deletion as skipped", func() {
					mockCloud.DeleteDiskReturns(bicloud.NewCPIError("delete_disk", bicloud.CmdError{
						Type:    bicloud.DiskNotFoundError,
						Message: "fake-disk-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.PerformCalls[0].Name).To(Equal("Deleting unused disk 'orphan-disk-cid'"))
					Expect(fakeStage.PerformCalls[0].SkipError.Error()).To(Equal("Disk Not Found: CPI 'delete_disk' method responded with error: CmdError{\"type\":\"Bosh::Clouds::DiskNotFound\",\"message\":\"fake-disk-not-found-message\",\"ok_to_retry\":false}"))
				})
			})
		})

		Context("orphan stemcell records exist", func() {
			BeforeEach(func() {
				_, err := stemcellRepo.Save("orphan-stemcell-name", "orphan-stemcell-version", "orphan-stemcell-cid", stemcellApiVersion)
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the unused stemcells", func() {
				mockCloud.DeleteStemcellReturns(nil)

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				stemcellRecords, err := stemcellRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
			})

			It("logs delete stage", func() {
				mockCloud.DeleteStemcellReturns(nil)

				err := deploymentManager.Cleanup(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(ContainElement(&testui.PerformCall{
					Name: "Deleting unused stemcell 'orphan-stemcell-cid'",
				}))
			})

			Context("when stemcells have been deleted manually (in the infrastructure)", func() {
				It("deletes the unused stemcells, ignoring StemcellNotFoundError", func() {
					mockCloud.DeleteStemcellReturns(bicloud.NewCPIError("delete_stemcell", bicloud.CmdError{
						Type:    bicloud.StemcellNotFoundError,
						Message: "fake-stemcell-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					stemcellRecords, err := diskRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
				})

				It("logs stemcell deletion as skipped", func() {
					mockCloud.DeleteStemcellReturns(bicloud.NewCPIError("delete_stemcell", bicloud.CmdError{
						Type:    bicloud.StemcellNotFoundError,
						Message: "fake-stemcell-not-found-message",
					}))

					err := deploymentManager.Cleanup(fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.PerformCalls[0].Name).To(Equal("Deleting unused stemcell 'orphan-stemcell-cid'"))
					Expect(fakeStage.PerformCalls[0].SkipError.Error()).To(Equal("Stemcell not found: CPI 'delete_stemcell' method responded with error: CmdError{\"type\":\"Bosh::Clouds::StemcellNotFound\",\"message\":\"fake-stemcell-not-found-message\",\"ok_to_retry\":false}"))
				})
			})
		})
	})
})
