package deployment_test

import (
	"time"

	"github.com/cloudfoundry/bosh-agent/v2/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/v2/agentclient/applyspec"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/agentclient/agentclientfakes"
	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	fakebiconfig "github.com/cloudfoundry/bosh-cli/v7/config/fakes"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state/statefakes"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	fakebisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel/fakes"
	fakebivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm/fakes"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/vm/vmfakes"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("Deployer", func() {
	var (
		deployer               Deployer
		mockVMManagerFactory   *vmfakes.FakeManagerFactory
		fakeVMManager          *fakebivm.FakeManager
		mockAgentClient        *agentclientfakes.FakeAgentClient
		mockAgentClientFactory *cmdfakes.FakeAgentClientFactory
		fakeSSHTunnelFactory   *fakebisshtunnel.FakeFactory
		fakeSSHTunnel          *fakebisshtunnel.FakeTunnel
		cloud                  *cloudfakes.FakeCloud
		deploymentManifest     bideplmanifest.Manifest
		diskPool               bideplmanifest.DiskPool
		fakeStage              *fakebiui.FakeStage
		fakeVM                 *fakebivm.FakeVM
		skipDrain              bool
		diskCIDs               []string

		cloudStemcell bistemcell.CloudStemcell

		applySpec bias.ApplySpec

		mockStateBuilderFactory *statefakes.FakeBuilderFactory
		mockStateBuilder        *statefakes.FakeBuilder
		mockState               *statefakes.FakeState

		mockBlobstore *blobstorefakes.FakeBlobstore
	)

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

		skipDrain = false
		diskCIDs = []string{"fake-disk-cid-1"}
		cloud = &cloudfakes.FakeCloud{}

		mockAgentClientFactory = &cmdfakes.FakeAgentClientFactory{}
		mockAgentClient = &agentclientfakes.FakeAgentClient{}
		mockAgentClientFactory.NewAgentClientReturns(mockAgentClient, nil)

		mockVMManagerFactory = &vmfakes.FakeManagerFactory{}
		fakeVMManager = fakebivm.NewFakeManager()
		mockVMManagerFactory.NewManagerReturns(fakeVMManager)

		fakeSSHTunnelFactory = fakebisshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebisshtunnel.NewFakeTunnel()
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel
		fakeSSHTunnel.SetStartBehavior(nil, nil)

		fakeVM = fakebivm.NewFakeVM("fake-vm-cid")
		fakeVMManager.CreateVM = fakeVM

		fakeVM.AgentClientReturn = mockAgentClient

		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeStage = fakebiui.NewFakeStage()

		fakeStemcellRepo := fakebiconfig.NewFakeStemcellRepo()
		stemcellRecord := biconfig.StemcellRecord{
			ID:      "fake-stemcell-id",
			Name:    "fake-stemcell-name",
			Version: "fake-stemcell-version",
			CID:     "fake-stemcell-cid",
		}
		err := fakeStemcellRepo.SetFindBehavior("fake-stemcell-name", "fake-stemcell-version", stemcellRecord, true, nil)
		Expect(err).ToNot(HaveOccurred())

		cloudStemcell = bistemcell.NewCloudStemcell(stemcellRecord, fakeStemcellRepo, cloud)

		mockStateBuilderFactory = &statefakes.FakeBuilderFactory{}
		mockStateBuilder = &statefakes.FakeBuilder{}
		mockState = &statefakes.FakeState{}

		instanceFactory := biinstance.NewFactory(mockStateBuilderFactory)
		instanceManagerFactory := biinstance.NewManagerFactory(fakeSSHTunnelFactory, instanceFactory, logger)

		mockBlobstore = &blobstorefakes.FakeBlobstore{}

		pingTimeout := 10 * time.Second
		pingDelay := 500 * time.Millisecond
		deploymentFactory := NewFactory(pingTimeout, pingDelay)

		deployer = NewDeployer(
			mockVMManagerFactory,
			instanceManagerFactory,
			deploymentFactory,
			logger,
		)
	})

	JustBeforeEach(func() {
		// since we're just passing this from State.ToApplySpec() to VM.Apply(), it doesn't need to be filled out
		applySpec = bias.ApplySpec{
			Deployment: "fake-deployment-name",
		}

		fakeAgentState := agentclient.AgentState{}
		fakeVM.GetStateResult = fakeAgentState

		mockStateBuilderFactory.NewBuilderReturns(mockStateBuilder)
		mockStateBuilder.BuildReturns(mockState, nil)
		mockStateBuilder.BuildInitialStateReturns(mockState, nil)
		mockState.ToApplySpecReturns(applySpec)
	})

	Context("when a previous instance exists", func() {
		var fakeExistingVM *fakebivm.FakeVM

		BeforeEach(func() {
			fakeExistingVM = fakebivm.NewFakeVM("existing-vm-cid")
			fakeVMManager.SetFindCurrentBehavior(fakeExistingVM, true, nil)
			fakeExistingVM.AgentClientReturn = mockAgentClient
		})

		It("deletes existing vm", func() {
			_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExistingVM.DeleteCalled).To(Equal(1))

			Expect(fakeStage.PerformCalls[:6]).To(Equal([]*fakebiui.PerformCall{
				{Name: "Waiting for the agent on VM 'existing-vm-cid'"},
				{Name: "Running the pre-stop scripts 'unknown/0'"},
				{Name: "Draining jobs on instance 'unknown/0'"},
				{Name: "Stopping jobs on instance 'unknown/0'"},
				{Name: "Running the post-stop scripts 'unknown/0'"},
				{Name: "Deleting VM 'existing-vm-cid'"},
			}))
		})

		Context("when skip-drain is specified", func() {
			It("skips draining", func() {
				skipDrain = true
				_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeExistingVM.DeleteCalled).To(Equal(1))

				Expect(fakeStage.PerformCalls[:5]).To(Equal([]*fakebiui.PerformCall{
					{Name: "Waiting for the agent on VM 'existing-vm-cid'"},
					{Name: "Running the pre-stop scripts 'unknown/0'"},
					{Name: "Stopping jobs on instance 'unknown/0'"},
					{Name: "Running the post-stop scripts 'unknown/0'"},
					{Name: "Deleting VM 'existing-vm-cid'"},
				}))
			})
		})
	})

	It("creates a vm", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVMManager.CreateInput).To(Equal(fakebivm.CreateInput{
			Stemcell: cloudStemcell,
			Manifest: deploymentManifest,
			DiskCIDs: diskCIDs,
		}))
	})

	It("waits for the vm", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebivm.WaitUntilReadyInput{
			Timeout: 10 * time.Minute,
			Delay:   500 * time.Millisecond,
		}))
	})

	It("logs start and stop events to the eventLogger", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStage.PerformCalls[1]).To(Equal(&fakebiui.PerformCall{
			Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready",
		}))
	})

	Context("when waiting for the agent fails", func() {
		var (
			waitError = bosherr.Error("fake-wait-error")
		)

		BeforeEach(func() {
			fakeVM.WaitUntilReadyErr = waitError
		})

		It("logs start and stop events to the eventLogger", func() {
			_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

			Expect(fakeStage.PerformCalls[1]).To(Equal(&fakebiui.PerformCall{
				Name:  "Waiting for the agent on VM 'fake-vm-cid' to be ready",
				Error: waitError,
			}))
		})
	})

	It("updates the vm", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.ApplyInputs).To(Equal([]fakebivm.ApplyInput{
			{ApplySpec: applySpec},
			{ApplySpec: applySpec},
		}))
	})

	It("starts the agent", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.StartCalled).To(Equal(1))
	})

	It("waits until agent reports state as running", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebivm.WaitInput{
			MaxAttempts: 5,
			Delay:       1 * time.Second,
		}))
	})

	Context("when the deployment has an invalid disk pool specification", func() {
		BeforeEach(func() {
			deploymentManifest.Jobs[0].PersistentDiskPool = "fake-non-existent-persistent-disk-pool-name"
		})

		It("returns an error", func() {
			_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
			Expect(err).To(HaveOccurred())
		})
	})

	It("logs instance update ui stages", func() {
		_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStage.PerformCalls[2:4]).To(Equal([]*fakebiui.PerformCall{
			{Name: "Updating instance 'fake-job-name/0'"},
			{Name: "Waiting for instance 'fake-job-name/0' to be running"},
		}))
	})

	Context("when applying instance spec fails", func() {
		BeforeEach(func() {
			fakeVM.ApplyErr = bosherr.Error("fake-apply-error")
		})

		It("fails with descriptive error", func() {
			_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Applying the initial agent state: fake-apply-error"))
		})
	})

	Context("when starting agent services fails", func() {
		BeforeEach(func() {
			fakeVM.StartErr = bosherr.Error("fake-start-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-start-error"))

			Expect(fakeStage.PerformCalls[2].Name).To(Equal("Updating instance 'fake-job-name/0'"))
			Expect(fakeStage.PerformCalls[2].Error).To(HaveOccurred())
			Expect(fakeStage.PerformCalls[2].Error.Error()).To(Equal("Starting the agent: fake-start-error"))
		})
	})

	Context("when waiting for running state fails", func() {
		var (
			waitError = bosherr.Error("fake-wait-running-error")
		)

		BeforeEach(func() {
			fakeVM.WaitToBeRunningErr = waitError
		})

		It("logs start and stop events to the eventLogger", func() {
			_, err := deployer.Deploy(cloud, deploymentManifest, cloudStemcell, fakeVMManager, mockBlobstore, skipDrain, diskCIDs, fakeStage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

			Expect(fakeStage.PerformCalls[3]).To(Equal(&fakebiui.PerformCall{
				Name:  "Waiting for instance 'fake-job-name/0' to be running",
				Error: waitError,
			}))
		})
	})
})
