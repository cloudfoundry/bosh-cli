package instance_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry/bosh-agent/v2/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/v2/agentclient/applyspec"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/agentclient/agentclientfakes"
	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/disk/diskfakes"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state/statefakes"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel/sshtunnelfakes"
	fakebivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm/fakes"
	fakebistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell/stemcellfakes"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("Manager", func() {
	var (
		apiVersion = 2
	)

	var (
		fakeCloud *cloudfakes.FakeCloud

		mockStateBuilderFactory *statefakes.FakeBuilderFactory
		mockStateBuilder        *statefakes.FakeBuilder
		mockState               *statefakes.FakeState

		mockBlobstore *blobstorefakes.FakeBlobstore

		fakeVMManager        *fakebivm.FakeManager
		fakeSSHTunnelFactory *sshtunnelfakes.FakeFactory
		fakeSSHTunnel        *sshtunnelfakes.FakeSSHTunnel
		instanceFactory      Factory
		logger               boshlog.Logger
		fakeStage            *fakebiui.FakeStage

		manager Manager
	)

	BeforeEach(func() {
		fakeCloud = &cloudfakes.FakeCloud{}
		fakeVMManager = fakebivm.NewFakeManager()

		fakeSSHTunnel = &sshtunnelfakes.FakeSSHTunnel{
			StartStub: func(readyErrCh chan<- error, errCh chan<- error) {
				readyErrCh <- nil
				errCh <- nil
			},
		}
		fakeSSHTunnelFactory = &sshtunnelfakes.FakeFactory{}
		fakeSSHTunnelFactory.NewSSHTunnelReturns(fakeSSHTunnel)

		mockStateBuilderFactory = &statefakes.FakeBuilderFactory{}
		mockStateBuilder = &statefakes.FakeBuilder{}
		mockState = &statefakes.FakeState{}

		instanceFactory = NewFactory(mockStateBuilderFactory)

		mockBlobstore = &blobstorefakes.FakeBlobstore{}

		logger = boshlog.NewLogger(boshlog.LevelNone)

		fakeStage = fakebiui.NewFakeStage()

		manager = NewManager(
			fakeCloud,
			fakeVMManager,
			mockBlobstore,
			fakeSSHTunnelFactory,
			instanceFactory,
			logger,
		)
	})

	Describe("Create", func() {
		var (
			mockAgentClient    *agentclientfakes.FakeAgentClient
			fakeVM             *fakebivm.FakeVM
			diskPool           bideplmanifest.DiskPool
			deploymentManifest bideplmanifest.Manifest
			fakeCloudStemcell  *fakebistemcell.FakeCloudStemcell
			diskCIDs           []string

			expectedInstance Instance
			expectedDisk     *diskfakes.FakeDisk
		)

		var allowApplySpecToBeCreated = func() {
			jobName := "cpi"
			jobIndex := 0

			applySpec := bias.ApplySpec{
				Deployment: "test-release",
				Index:      jobIndex,
				Packages:   map[string]bias.Blob{},
				Networks: map[string]interface{}{
					"network-1": biproperty.Map{
						"cloud_properties": biproperty.Map{},
						"type":             "dynamic",
						"ip":               "",
					},
				},
				Job: bias.Job{
					Name:      jobName,
					Templates: []bias.Blob{},
				},
				RenderedTemplatesArchive: bias.RenderedTemplatesArchiveSpec{},
				ConfigurationHash:        "",
			}

			fakeAgentState := agentclient.AgentState{}
			fakeVM.GetStateResult = fakeAgentState

			mockStateBuilderFactory.NewBuilderReturns(mockStateBuilder)
			mockStateBuilder.BuildReturns(mockState, nil)
			mockState.ToApplySpecReturns(applySpec)
		}

		BeforeEach(func() {
			diskPool = bideplmanifest.DiskPool{
				Name:     "fake-persistent-disk-pool-name",
				DiskSize: 1024,
				CloudProperties: biproperty.Map{
					"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
				},
			}

			deploymentManifest = bideplmanifest.Manifest{
				Update: bideplmanifest.Update{
					UpdateWatchTime: bideplmanifest.WatchTime{
						Start: 0,
						End:   5478,
					},
				},
				DiskPools: []bideplmanifest.DiskPool{
					diskPool,
				},
				Jobs: []bideplmanifest.Job{
					{
						Name:               "fake-job-name",
						PersistentDiskPool: "fake-persistent-disk-pool-name",
						Instances:          1,
					},
				},
			}

			fakeCloudStemcell = fakebistemcell.NewFakeCloudStemcell("fake-stemcell-cid", "fake-stemcell-name", "fake-stemcell-version", apiVersion)

			diskCIDs = []string{"fake-disk-cid"}

			fakeVM = fakebivm.NewFakeVM("fake-vm-cid")
			fakeVMManager.CreateVM = fakeVM

			mockAgentClient = &agentclientfakes.FakeAgentClient{}
			fakeVM.AgentClientReturn = mockAgentClient

			expectedInstance = NewInstance(
				"fake-job-name",
				0,
				fakeVM,
				fakeVMManager,
				fakeSSHTunnelFactory,
				mockStateBuilder,
				logger,
			)

			expectedDisk = &diskfakes.FakeDisk{}
			fakeVM.UpdateDisksDisks = []bidisk.Disk{expectedDisk}
		})

		JustBeforeEach(func() {
			allowApplySpecToBeCreated()
		})

		It("returns an Instance that wraps a newly created VM", func() {
			instance, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				diskCIDs,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).To(Equal(expectedInstance))

			Expect(fakeVMManager.CreateInput).To(Equal(fakebivm.CreateInput{
				Stemcell: fakeCloudStemcell,
				Manifest: deploymentManifest,
				DiskCIDs: diskCIDs,
			}))
		})

		It("updates the current stemcell", func() {
			_, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				diskCIDs,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCloudStemcell.PromoteAsCurrentCalledTimes).To(Equal(1))
		})

		It("logs instance update stages", func() {
			_, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				diskCIDs,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Creating VM for instance 'fake-job-name/0' from stemcell 'fake-stemcell-cid'"},
				{Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready"},
			}))
		})

		It("waits for the vm", func() {
			_, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				diskCIDs,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.WaitUntilReadyInputs).To(Equal([]fakebivm.WaitUntilReadyInput{
				{
					Timeout: 10 * time.Minute,
					Delay:   500 * time.Millisecond,
				},
			}))

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Creating VM for instance 'fake-job-name/0' from stemcell 'fake-stemcell-cid'"},
				{Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready"},
			}))
		})

		It("returns the 'updated' disks", func() {
			_, disks, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				diskCIDs,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(disks).To(Equal([]bidisk.Disk{expectedDisk}))

			Expect(fakeVM.UpdateDisksInputs).To(Equal([]fakebivm.UpdateDisksInput{
				{
					DiskPool: diskPool,
					Stage:    fakeStage,
				},
			}))
		})

		Context("when creating VM fails", func() {
			BeforeEach(func() {
				fakeVMManager.CreateErr = errors.New("fake-create-vm-error")
			})

			It("returns an error", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					diskCIDs,
					fakeStage,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					diskCIDs,
					fakeStage,
				)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Creating VM for instance 'fake-job-name/0' from stemcell 'fake-stemcell-cid'"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Creating VM: fake-create-vm-error"))
			})
		})
	})
})
