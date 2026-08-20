package deployment_test

import (
	"time"

	biagentclient "github.com/cloudfoundry/bosh-agent/v2/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/v2/agentclient/applyspec"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
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
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state/statefakes"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("Deployment", func() {

	var (
		logger boshlog.Logger
		fs     boshsys.FileSystem

		fakeUUIDGenerator      *fakeuuid.FakeGenerator
		fakeRepoUUIDGenerator  *fakeuuid.FakeGenerator
		deploymentStateService biconfig.DeploymentStateService
		vmRepo                 biconfig.VMRepo
		diskRepo               biconfig.DiskRepo
		stemcellRepo           biconfig.StemcellRepo

		mockCloud       *cloudfakes.FakeCloud
		mockAgentClient *agentclientfakes.FakeAgentClient

		mockStateBuilderFactory *statefakes.FakeBuilderFactory
		mockStateBuilder        *statefakes.FakeBuilder
		mockState               *statefakes.FakeState

		mockBlobstore *blobstorefakes.FakeBlobstore

		fakeStage *fakebiui.FakeStage

		deploymentFactory Factory

		stemcellApiVersion = 2
		deployment         Deployment
		skipDrain          bool

		// callOrder records, in invocation order, every tracked mockCloud/mockAgentClient
		// call made during a test -- the counterfeiter equivalent of gomock.InOrder().
		callOrder []string
	)

	var allowApplySpecToBeCreated = func() {
		jobIndex := 0

		applySpec := bias.ApplySpec{
			Deployment: "test-release",
			Index:      jobIndex,
			Packages:   map[string]bias.Blob{},
			Networks: map[string]interface{}{
				"network-1": map[string]interface{}{
					"cloud_properties": map[string]interface{}{},
					"type":             "dynamic",
					"ip":               "",
				},
			},
			Job: bias.Job{
				Name:      "fake-job-name",
				Templates: []bias.Blob{},
			},
			RenderedTemplatesArchive: bias.RenderedTemplatesArchiveSpec{},
			ConfigurationHash:        "",
		}

		mockStateBuilderFactory.NewBuilderReturns(mockStateBuilder)
		mockState.ToApplySpecReturns(applySpec)
	}

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()

		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		deploymentStateService = biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, "/deployment.json")

		fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()
		vmRepo = biconfig.NewVMRepo(deploymentStateService)
		diskRepo = biconfig.NewDiskRepo(deploymentStateService, fakeRepoUUIDGenerator)
		stemcellRepo = biconfig.NewStemcellRepo(deploymentStateService, fakeRepoUUIDGenerator)

		callOrder = nil

		mockCloud = &cloudfakes.FakeCloud{}
		mockCloud.HasVMStub = func(cid string) (bool, error) {
			callOrder = append(callOrder, "HasVM")
			return true, nil
		}
		mockCloud.DeleteVMStub = func(cid string) error {
			callOrder = append(callOrder, "DeleteVM")
			return nil
		}
		mockCloud.DeleteDiskStub = func(cid string) error {
			callOrder = append(callOrder, "DeleteDisk")
			return nil
		}
		mockCloud.DeleteStemcellStub = func(cid string) error {
			callOrder = append(callOrder, "DeleteStemcell")
			return nil
		}

		mockAgentClient = &agentclientfakes.FakeAgentClient{}
		mockAgentClient.PingStub = func() (string, error) {
			callOrder = append(callOrder, "Ping")
			return "any-state", nil
		}
		mockAgentClient.RunScriptStub = func(name string, opts map[string]interface{}) error {
			callOrder = append(callOrder, "RunScript:"+name)
			return nil
		}
		mockAgentClient.DrainStub = func(state string) (int64, error) {
			callOrder = append(callOrder, "Drain")
			return 0, nil
		}
		mockAgentClient.StopStub = func() error {
			callOrder = append(callOrder, "Stop")
			return nil
		}
		mockAgentClient.ListDiskStub = func() ([]string, error) {
			callOrder = append(callOrder, "ListDisk")
			return []string{"fake-disk-cid"}, nil
		}
		mockAgentClient.UnmountDiskStub = func(cid string) error {
			callOrder = append(callOrder, "UnmountDisk")
			return nil
		}
		mockAgentClient.StartStub = func() error {
			callOrder = append(callOrder, "Start")
			return nil
		}
		mockAgentClient.GetStateStub = func() (biagentclient.AgentState, error) {
			callOrder = append(callOrder, "GetState")
			return biagentclient.AgentState{JobState: "running"}, nil
		}

		fakeStage = fakebiui.NewFakeStage()

		pingTimeout := 10 * time.Second
		pingDelay := 500 * time.Millisecond
		deploymentFactory = NewFactory(pingTimeout, pingDelay)

		skipDrain = false
	})

	JustBeforeEach(func() {
		// all these local factories & managers are just used to construct a Deployment based on the deployment state
		diskManagerFactory := bidisk.NewManagerFactory(diskRepo, logger)
		diskDeployer := bivm.NewDiskDeployer(diskManagerFactory, diskRepo, logger, false)

		vmManagerFactory := bivm.NewManagerFactory(vmRepo, stemcellRepo, diskDeployer, fakeUUIDGenerator, fs, logger)
		sshTunnelFactory := bisshtunnel.NewFactory(logger)

		mockStateBuilderFactory = &statefakes.FakeBuilderFactory{}
		mockStateBuilder = &statefakes.FakeBuilder{}
		mockState = &statefakes.FakeState{}

		instanceFactory := biinstance.NewFactory(mockStateBuilderFactory)
		instanceManagerFactory := biinstance.NewManagerFactory(sshTunnelFactory, instanceFactory, logger)
		stemcellManagerFactory := bistemcell.NewManagerFactory(stemcellRepo)

		mockBlobstore = &blobstorefakes.FakeBlobstore{}

		deploymentManagerFactory := NewManagerFactory(vmManagerFactory, instanceManagerFactory, diskManagerFactory, stemcellManagerFactory, deploymentFactory)
		deploymentManager := deploymentManagerFactory.NewManager(mockCloud, mockAgentClient, mockBlobstore)

		allowApplySpecToBeCreated()

		var err error
		deployment, _, err = deploymentManager.FindCurrent()
		Expect(err).ToNot(HaveOccurred())
		// Note: deployment will be nil if the config has no vms, disks, or stemcells
	})

	Describe("Delete", func() {

		normalFlowOrder := []string{
			"HasVM", "Ping", "RunScript:pre-stop", "Drain", "Stop", "RunScript:post-stop",
			"ListDisk", "UnmountDisk", "DeleteVM", "DeleteDisk", "DeleteStemcell",
		}
		drainlessFlowOrder := []string{
			"HasVM", "Ping", "RunScript:pre-stop", "Stop", "RunScript:post-stop",
			"ListDisk", "UnmountDisk", "DeleteVM", "DeleteDisk", "DeleteStemcell",
		}

		Context("when the deployment has been deployed", func() {
			BeforeEach(func() {
				// create deployment manifest yaml file
				err := deploymentStateService.Save(biconfig.DeploymentState{
					DirectorID:        "fake-director-id",
					InstallationID:    "fake-installation-id",
					CurrentVMCID:      "fake-vm-cid",
					CurrentStemcellID: "fake-stemcell-guid",
					CurrentDiskID:     "fake-disk-guid",
					Disks: []biconfig.DiskRecord{
						{
							ID:   "fake-disk-guid",
							CID:  "fake-disk-cid",
							Size: 100,
						},
					},
					Stemcells: []biconfig.StemcellRecord{
						{
							ID:  "fake-stemcell-guid",
							CID: "fake-stemcell-cid",
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			It("stops agent, unmounts disk, deletes vm, deletes disk, deletes stemcell", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(callOrder).To(Equal(normalFlowOrder))
			})

			It("skips draining if specified", func() {
				skipDrain = true

				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(callOrder).To(Equal(drainlessFlowOrder))
			})

			It("logs validation stages", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
					{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
					{Name: "Running the pre-stop scripts 'unknown/0'"},
					{Name: "Draining jobs on instance 'unknown/0'"},
					{Name: "Stopping jobs on instance 'unknown/0'"},
					{Name: "Running the post-stop scripts 'unknown/0'"},
					{Name: "Unmounting disk 'fake-disk-cid'"},
					{Name: "Deleting VM 'fake-vm-cid'"},
					{Name: "Deleting disk 'fake-disk-cid'"},
					{Name: "Deleting stemcell 'fake-stemcell-cid'"},
				}))
			})

			It("clears current vm, disk and stemcell", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				_, found, err := vmRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse(), "should be no current VM")

				_, found, err = diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse(), "should be no current disk")

				diskRecords, err := diskRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(diskRecords).To(BeEmpty(), "expected no disk records")

				_, found, err = stemcellRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse(), "should be no current stemcell")

				stemcellRecords, err := stemcellRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
			})

			// TODO: It'd be nice to test recovering after agent was responsive, before timeout (hard to do with gomock)
			Context("when agent is unresponsive", func() {
				BeforeEach(func() {
					// reduce timout & delay to reduce test duration
					pingTimeout := 1 * time.Second
					pingDelay := 100 * time.Millisecond
					deploymentFactory = NewFactory(pingTimeout, pingDelay)
				})

				It("times out pinging agent, deletes vm, deletes disk, deletes stemcell", func() {
					mockAgentClient.PingStub = func() (string, error) {
						callOrder = append(callOrder, "Ping")
						return "", bosherr.Error("unresponsive agent")
					}

					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("and delete previously suceeded", func() {
				JustBeforeEach(func() {
					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					// reset event log recording
					fakeStage = fakebiui.NewFakeStage()
				})

				It("does not delete anything", func() {
					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.PerformCalls).To(BeEmpty())
				})
			})
		})

		Context("when nothing has been deployed", func() {
			BeforeEach(func() {
				err := deploymentStateService.Save(biconfig.DeploymentState{})
				Expect(err).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				// A previous JustBeforeEach uses FindCurrent to define deployment,
				// which would return a nil if the config is empty.
				// So we have to make a fake empty deployment to test it.
				deployment = deploymentFactory.NewDeployment([]biinstance.Instance{}, []bidisk.Disk{}, []bistemcell.CloudStemcell{})
			})

			It("does not delete anything", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(BeEmpty())
			})
		})

		Context("when VM has been deployed", func() {
			BeforeEach(func() {
				err := deploymentStateService.Save(biconfig.DeploymentState{})
				Expect(err).ToNot(HaveOccurred())
				err = vmRepo.UpdateCurrent("fake-vm-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("stops the agent and deletes the VM", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(callOrder).To(Equal([]string{
					"HasVM", "Ping", "RunScript:pre-stop", "Drain", "Stop", "RunScript:post-stop",
					"ListDisk", "UnmountDisk", "DeleteVM",
				}))
			})

			Context("when VM has been deleted manually (outside of bosh)", func() {
				BeforeEach(func() {
					mockCloud.HasVMStub = func(cid string) (bool, error) {
						callOrder = append(callOrder, "HasVM")
						return false, nil
					}
				})

				It("skips agent shutdown & deletes the VM (to ensure related resources are released by the CPI)", func() {
					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(mockCloud.DeleteVMCallCount()).To(Equal(1))
				})

				It("ignores VMNotFound errors", func() {
					mockCloud.DeleteVMStub = func(cid string) error {
						callOrder = append(callOrder, "DeleteVM")
						return bicloud.NewCPIError("delete_vm", bicloud.CmdError{
							Type:    bicloud.VMNotFoundError,
							Message: "fake-vm-not-found-message",
						})
					}

					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when a current disk exists", func() {
			BeforeEach(func() {
				err := deploymentStateService.Save(biconfig.DeploymentState{})
				Expect(err).ToNot(HaveOccurred())
				diskRecord, err := diskRepo.Save("fake-disk-cid", 100, nil)
				Expect(err).ToNot(HaveOccurred())
				err = diskRepo.UpdateCurrent(diskRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the disk", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(mockCloud.DeleteDiskCallCount()).To(Equal(1))
			})

			Context("when current disk has been deleted manually (outside of bosh)", func() {
				It("deletes the disk (to ensure related resources are released by the CPI)", func() {
					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(mockCloud.DeleteDiskCallCount()).To(Equal(1))
				})

				It("ignores DiskNotFound errors", func() {
					mockCloud.DeleteDiskStub = func(cid string) error {
						callOrder = append(callOrder, "DeleteDisk")
						return bicloud.NewCPIError("delete_disk", bicloud.CmdError{
							Type:    bicloud.DiskNotFoundError,
							Message: "fake-disk-not-found-message",
						})
					}

					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when a current stemcell exists", func() {
			BeforeEach(func() {
				err := deploymentStateService.Save(biconfig.DeploymentState{})
				Expect(err).ToNot(HaveOccurred())
				stemcellRecord, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-stemcell-cid", stemcellApiVersion)
				Expect(err).ToNot(HaveOccurred())
				err = stemcellRepo.UpdateCurrent(stemcellRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the stemcell", func() {
				err := deployment.Delete(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(mockCloud.DeleteStemcellCallCount()).To(Equal(1))
			})

			Context("when current stemcell has been deleted manually (outside of bosh)", func() {
				It("deletes the stemcell (to ensure related resources are released by the CPI)", func() {
					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(mockCloud.DeleteStemcellCallCount()).To(Equal(1))
				})

				It("ignores StemcellNotFound errors", func() {
					mockCloud.DeleteStemcellStub = func(cid string) error {
						callOrder = append(callOrder, "DeleteStemcell")
						return bicloud.NewCPIError("delete_stemcell", bicloud.CmdError{
							Type:    bicloud.StemcellNotFoundError,
							Message: "fake-stemcell-not-found-message",
						})
					}

					err := deployment.Delete(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Stop", func() {

		stopNormalFlowOrder := []string{"Ping", "RunScript:pre-stop", "Drain", "Stop", "RunScript:post-stop"}
		stopDrainlessFlowOrder := []string{"Ping", "RunScript:pre-stop", "Stop", "RunScript:post-stop"}

		Context("when the deployment has been deployed", func() {
			BeforeEach(func() {
				// create deployment manifest yaml file
				err := deploymentStateService.Save(biconfig.DeploymentState{
					DirectorID:        "fake-director-id",
					InstallationID:    "fake-installation-id",
					CurrentVMCID:      "fake-vm-cid",
					CurrentStemcellID: "fake-stemcell-guid",
					CurrentDiskID:     "fake-disk-guid",
					Disks: []biconfig.DiskRecord{
						{
							ID:   "fake-disk-guid",
							CID:  "fake-disk-cid",
							Size: 100,
						},
					},
					Stemcells: []biconfig.StemcellRecord{
						{
							ID:  "fake-stemcell-guid",
							CID: "fake-stemcell-cid",
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			It("stops agent and executes the pre-stop and post-stop scripts", func() {
				err := deployment.Stop(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(callOrder).To(Equal(stopNormalFlowOrder))
			})

			It("skips draining if specified", func() {
				skipDrain = true

				err := deployment.Stop(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(callOrder).To(Equal(stopDrainlessFlowOrder))
			})

			It("logs validation stages", func() {
				err := deployment.Stop(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
					{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
					{Name: "Running the pre-stop scripts 'unknown/0'"},
					{Name: "Draining jobs on instance 'unknown/0'"},
					{Name: "Stopping jobs on instance 'unknown/0'"},
					{Name: "Running the post-stop scripts 'unknown/0'"},
				}))
			})

			Context("when agent is unresponsive", func() {
				BeforeEach(func() {
					// reduce timout & delay to reduce test duration
					pingTimeout := 1 * time.Second
					pingDelay := 100 * time.Millisecond
					deploymentFactory = NewFactory(pingTimeout, pingDelay)
				})

				It("times out pinging agent and does nothing", func() {
					mockAgentClient.PingStub = func() (string, error) {
						callOrder = append(callOrder, "Ping")
						return "", bosherr.Error("unresponsive agent")
					}

					err := deployment.Stop(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("and delete previously suceeded", func() {
				JustBeforeEach(func() {
					err := deployment.Stop(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					// reset event log recording
					fakeStage = fakebiui.NewFakeStage()
				})

				It("does not delete anything", func() {
					err := deployment.Stop(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.PerformCalls).To(BeEmpty())
				})
			})
		})

		Context("when nothing has been deployed", func() {
			BeforeEach(func() {
				err := deploymentStateService.Save(biconfig.DeploymentState{})
				Expect(err).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				// A previous JustBeforeEach uses FindCurrent to define deployment,
				// which would return a nil if the config is empty.
				// So we have to make a fake empty deployment to test it.
				deployment = deploymentFactory.NewDeployment([]biinstance.Instance{}, []bidisk.Disk{}, []bistemcell.CloudStemcell{})
			})

			It("does not stop anything", func() {
				err := deployment.Stop(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(BeEmpty())
			})
		})
	})

	Describe("Start", func() {
		startNormalFlowOrder := []string{"Ping", "RunScript:pre-start", "Start", "GetState", "RunScript:post-start"}

		var update = bideplmanifest.Update{
			UpdateWatchTime: bideplmanifest.WatchTime{
				Start: 0,
				End:   5478,
			},
		}

		Context("when the deployment has been deployed", func() {
			BeforeEach(func() {
				// create deployment manifest yaml file
				err := deploymentStateService.Save(biconfig.DeploymentState{
					DirectorID:        "fake-director-id",
					InstallationID:    "fake-installation-id",
					CurrentVMCID:      "fake-vm-cid",
					CurrentStemcellID: "fake-stemcell-guid",
					CurrentDiskID:     "fake-disk-guid",
					Disks: []biconfig.DiskRecord{
						{
							ID:   "fake-disk-guid",
							CID:  "fake-disk-cid",
							Size: 100,
						},
					},
					Stemcells: []biconfig.StemcellRecord{
						{
							ID:  "fake-stemcell-guid",
							CID: "fake-stemcell-cid",
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())
			})

			It("starts agent and executes the pre-start and post-start scripts", func() {
				err := deployment.Start(fakeStage, update)
				Expect(err).ToNot(HaveOccurred())
				Expect(callOrder).To(Equal(startNormalFlowOrder))
			})

			It("logs validation stages", func() {
				err := deployment.Start(fakeStage, update)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
					{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
					{Name: "Running the pre-start scripts 'unknown/0'"},
					{Name: "Starting the agent 'unknown/0'"},
					{Name: "Waiting for instance 'unknown/0' to be running"},
					{Name: "Running the post-start scripts 'unknown/0'"},
				}))
			})

			Context("when agent is unresponsive", func() {
				BeforeEach(func() {
					// reduce timout & delay to reduce test duration
					pingTimeout := 1 * time.Second
					pingDelay := 100 * time.Millisecond
					deploymentFactory = NewFactory(pingTimeout, pingDelay)
				})

				It("times out pinging agent and does nothing", func() {
					mockAgentClient.PingStub = func() (string, error) {
						callOrder = append(callOrder, "Ping")
						return "", bosherr.Error("unresponsive agent")
					}

					err := deployment.Start(fakeStage, update)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("and start previously suceeded", func() {
				JustBeforeEach(func() {
					err := deployment.Start(fakeStage, update)
					Expect(err).ToNot(HaveOccurred())

					// reset event log recording
					fakeStage = fakebiui.NewFakeStage()
					callOrder = nil
				})

				It("does execute the normal flow", func() {
					err := deployment.Start(fakeStage, update)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when nothing has been deployed", func() {
			BeforeEach(func() {
				err := deploymentStateService.Save(biconfig.DeploymentState{})
				Expect(err).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				// A previous JustBeforeEach uses FindCurrent to define deployment,
				// which would return a nil if the config is empty.
				// So we have to make a fake empty deployment to test it.
				deployment = deploymentFactory.NewDeployment([]biinstance.Instance{}, []bidisk.Disk{}, []bistemcell.CloudStemcell{})
			})

			It("does not start anything", func() {
				err := deployment.Start(fakeStage, update)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(BeEmpty())
			})
		})
	})
})
