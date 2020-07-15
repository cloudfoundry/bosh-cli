package instance_test

import (
	. "github.com/cloudfoundry/bosh-cli/deployment/instance"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mock_instance_state "github.com/cloudfoundry/bosh-cli/deployment/instance/state/mocks"
	"github.com/golang/mock/gomock"

	bias "github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/logger/loggerfakes"

	"github.com/cloudfoundry/bosh-agent/agentclient"
	fakebidisk "github.com/cloudfoundry/bosh-cli/deployment/disk/fakes"
	fakebisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel/fakes"
	fakebivm "github.com/cloudfoundry/bosh-cli/deployment/vm/fakes"
	fakebiui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("Instance", func() {

	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		mockStateBuilder *mock_instance_state.MockBuilder
		mockState        *mock_instance_state.MockState

		fakeVMManager        *fakebivm.FakeManager
		fakeVM               *fakebivm.FakeVM
		fakeSSHTunnelFactory *fakebisshtunnel.FakeFactory
		fakeSSHTunnel        *fakebisshtunnel.FakeTunnel
		fakeStage            *fakebiui.FakeStage
		logger               *loggerfakes.FakeLogger

		instance Instance

		pingTimeout = 1 * time.Second
		pingDelay   = 500 * time.Millisecond

		skipDrain bool

		jobName  = "fake-job-name"
		jobIndex = 0
	)

	BeforeEach(func() {
		fakeVMManager = fakebivm.NewFakeManager()
		fakeVM = fakebivm.NewFakeVM("fake-vm-cid")

		fakeSSHTunnelFactory = fakebisshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebisshtunnel.NewFakeTunnel()
		fakeSSHTunnel.SetStartBehavior(nil, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel

		mockStateBuilder = mock_instance_state.NewMockBuilder(mockCtrl)
		mockState = mock_instance_state.NewMockState(mockCtrl)

		logger = &loggerfakes.FakeLogger{}

		skipDrain = false

		instance = NewInstance(
			jobName,
			jobIndex,
			fakeVM,
			fakeVMManager,
			fakeSSHTunnelFactory,
			mockStateBuilder,
			logger,
		)

		fakeStage = fakebiui.NewFakeStage()
	})

	Describe("Delete", func() {
		It("checks if the agent on the vm is responsive", func() {
			err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebivm.WaitUntilReadyInput{
				Timeout: pingTimeout,
				Delay:   pingDelay,
			}))
		})

		It("deletes existing vm", func() {
			err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.DeleteCalled).To(Equal(1))
		})

		It("logs start and stop events", func() {
			err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
				{Name: "Running the pre-stop scripts 'fake-job-name/0'"},
				{Name: "Draining jobs on instance 'fake-job-name/0'"},
				{Name: "Stopping jobs on instance 'fake-job-name/0'"},
				{Name: "Running the post-stop scripts 'fake-job-name/0'"},
				{Name: "Deleting VM 'fake-vm-cid'"},
			}))
		})

		Context("when agent is responsive", func() {
			It("logs waiting for the agent event", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0]).To(Equal(&fakebiui.PerformCall{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
				}))
			})

			It("drains", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.DrainCalled).To(Equal(1))
			})

			It("can skip draining", func() {
				skipDrain = true
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.DrainCalled).To(Equal(0))
			})

			It("stops vm", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.StopCalled).To(Equal(1))
			})

			It("unmounts vm disks", func() {
				firstDisk := fakebidisk.NewFakeDisk("fake-disk-1")
				secondDisk := fakebidisk.NewFakeDisk("fake-disk-2")
				fakeVM.ListDisksDisks = []bidisk.Disk{firstDisk, secondDisk}

				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.UnmountDiskInputs).To(Equal([]fakebivm.UnmountDiskInput{
					{Disk: firstDisk},
					{Disk: secondDisk},
				}))

				Expect(fakeStage.PerformCalls[5:7]).To(Equal([]*fakebiui.PerformCall{
					{Name: "Unmounting disk 'fake-disk-1'"},
					{Name: "Unmounting disk 'fake-disk-2'"},
				}))
			})

			Context("when stopping vm fails", func() {
				var (
					stopError = bosherr.Error("fake-stop-error")
				)

				BeforeEach(func() {
					fakeVM.StopErr = stopError
				})

				It("returns an error", func() {
					err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-stop-error"))

					Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
						{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
						{Name: "Running the pre-stop scripts 'fake-job-name/0'"},
						{Name: "Draining jobs on instance 'fake-job-name/0'"},
						{
							Name:  "Stopping jobs on instance 'fake-job-name/0'",
							Error: stopError,
						},
					}))
				})
			})

			Context("when unmounting disk fails", func() {
				BeforeEach(func() {
					fakeVM.ListDisksDisks = []bidisk.Disk{fakebidisk.NewFakeDisk("fake-disk")}
					fakeVM.UnmountDiskErr = bosherr.Error("fake-unmount-error")
				})

				It("returns an error", func() {
					err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-unmount-error"))

					Expect(fakeStage.PerformCalls[5].Name).To(Equal("Unmounting disk 'fake-disk'"))
					Expect(fakeStage.PerformCalls[5].Error).To(HaveOccurred())
					Expect(fakeStage.PerformCalls[5].Error.Error()).To(Equal("Unmounting disk 'fake-disk' from VM 'fake-vm-cid': fake-unmount-error"))
				})
			})
		})

		Context("when agent fails to respond", func() {
			BeforeEach(func() {
				fakeVM.WaitUntilReadyErr = bosherr.Error("fake-wait-error")
			})

			It("logs failed event", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Waiting for the agent on VM 'fake-vm-cid'"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Agent unreachable: fake-wait-error"))
			})
		})

		Context("when deleting VM fails", func() {
			var (
				deleteError = bosherr.Error("fake-delete-error")
			)
			BeforeEach(func() {
				fakeVM.DeleteErr = deleteError
			})

			It("returns an error", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))

				Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
					{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
					{Name: "Running the pre-stop scripts 'fake-job-name/0'"},
					{Name: "Draining jobs on instance 'fake-job-name/0'"},
					{Name: "Stopping jobs on instance 'fake-job-name/0'"},
					{Name: "Running the post-stop scripts 'fake-job-name/0'"},
					{
						Name:  "Deleting VM 'fake-vm-cid'",
						Error: deleteError,
					},
				}))
			})
		})

		Context("when VM does not exist (deleted manually)", func() {
			BeforeEach(func() {
				fakeVM.ExistsFound = false
				fakeVM.DeleteErr = bicloud.NewCPIError("delete_vm", bicloud.CmdError{
					Type:    bicloud.VMNotFoundError,
					Message: "fake-vm-not-found-message",
				})
			})

			It("deletes existing vm", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeVM.DeleteCalled).To(Equal(1))
			})

			It("does not contact the agent", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeVM.WaitUntilReadyInputs).To(HaveLen(0))
				Expect(fakeVM.StopCalled).To(Equal(0))
				Expect(fakeVM.UnmountDiskInputs).To(HaveLen(0))
			})

			It("logs vm delete as skipped", func() {
				err := instance.Delete(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Deleting VM 'fake-vm-cid'"))
				Expect(fakeStage.PerformCalls[0].SkipError.Error()).To(Equal("VM not found: CPI 'delete_vm' method responded with error: CmdError{\"type\":\"Bosh::Clouds::VMNotFound\",\"message\":\"fake-vm-not-found-message\",\"ok_to_retry\":false}"))
			})
		})
	})

	Describe("UpdateJobs", func() {
		var (
			deploymentManifest bideplmanifest.Manifest

			applySpec bias.ApplySpec

			expectStateBuild *gomock.Call

			expectStateBuildInitialState *gomock.Call
		)

		BeforeEach(func() {
			// manifest is only being used for the Update.UpdateWatchTime, otherwise it's just being passed through to the StateBuilder
			deploymentManifest = bideplmanifest.Manifest{
				Name: "fake-deployment-name",
				Update: bideplmanifest.Update{
					UpdateWatchTime: bideplmanifest.WatchTime{
						Start: 0,
						End:   5478,
					},
				},
			}

			// apply spec is just returned from instance.State.ToApplySpec() and passed to agentClient.Apply()
			applySpec = bias.ApplySpec{
				Deployment: "fake-deployment-name",
			}
		})

		JustBeforeEach(func() {
			fakeAgentState := agentclient.AgentState{JobState: "testing"}
			fakeVM.GetStateResult = fakeAgentState

			expectStateBuild = mockStateBuilder.EXPECT().Build(jobName, jobIndex, deploymentManifest, fakeStage, fakeAgentState).Return(mockState, nil).AnyTimes()
			expectStateBuildInitialState = mockStateBuilder.EXPECT().BuildInitialState(jobName, jobIndex, deploymentManifest).Return(mockState, nil).AnyTimes()
			mockState.EXPECT().ToApplySpec().Return(applySpec).AnyTimes()
		})

		It("builds a new instance state", func() {
			expectStateBuild.Times(1)
			expectStateBuildInitialState.Times(1)

			err := instance.UpdateJobs(deploymentManifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())
		})

		It("tells agent to stop jobs, apply a new spec (with new rendered jobs templates), and start jobs", func() {
			err := instance.UpdateJobs(deploymentManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.StopCalled).To(Equal(1))
			Expect(fakeVM.ApplyInputs).To(Equal([]fakebivm.ApplyInput{
				{ApplySpec: applySpec},
				{ApplySpec: applySpec},
			}))
			Expect(fakeVM.RunScriptInputs).To(Equal([]string{"pre-start", "post-start"}))
			Expect(fakeVM.StartCalled).To(Equal(1))
		})

		It("waits until agent reports state as running", func() {
			err := instance.UpdateJobs(deploymentManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebivm.WaitInput{
				MaxAttempts: 5,
				Delay:       1 * time.Second,
			}))
		})

		It("logs start and stop events to the eventLogger", func() {
			err := instance.UpdateJobs(deploymentManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Updating instance 'fake-job-name/0'"},
				{Name: "Waiting for instance 'fake-job-name/0' to be running"},
				{Name: "Running the post-start scripts 'fake-job-name/0'"},
			}))
		})

		Context("when instance state building fails", func() {
			JustBeforeEach(func() {
				expectStateBuild.Return(nil, bosherr.Error("fake-template-err")).Times(1)
			})

			It("returns an error", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-err"))
			})
		})

		Context("when stopping vm fails", func() {
			BeforeEach(func() {
				fakeVM.StopErr = bosherr.Error("fake-stop-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stop-error"))

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Updating instance 'fake-job-name/0'"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Stopping the agent: fake-stop-error"))
			})
		})

		Context("when applying a new vm state fails", func() {
			BeforeEach(func() {
				fakeVM.ApplyErr = bosherr.Error("fake-apply-error")
			})

			It("fails with descriptive error", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Applying the initial agent state: fake-apply-error"))
			})
		})

		Context("when running the pre-start script fails", func() {
			BeforeEach(func() {
				fakeVM.RunScriptErrors["pre-start"] = bosherr.Error("fake-run-script-error")
			})

			It("returns the error", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-script-error"))
			})
		})

		Context("when running the post-start script fails", func() {
			BeforeEach(func() {
				fakeVM.RunScriptErrors["post-start"] = bosherr.Error("fake-run-script-error-poststart")
			})

			It("returns the error", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-script-error-poststart"))
			})
		})

		Context("when starting vm fails", func() {
			BeforeEach(func() {
				fakeVM.StartErr = bosherr.Error("fake-start-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-start-error"))

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Updating instance 'fake-job-name/0'"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Starting the agent: fake-start-error"))
			})
		})

		Context("when waiting for running state fails", func() {
			var (
				waitError = bosherr.Error("fake-wait-running-error")
			)

			BeforeEach(func() {
				fakeVM.WaitToBeRunningErr = waitError
			})

			It("logs instance update stages", func() {
				err := instance.UpdateJobs(deploymentManifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

				Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
					{Name: "Updating instance 'fake-job-name/0'"},
					{
						Name:  "Waiting for instance 'fake-job-name/0' to be running",
						Error: waitError,
					},
				}))
			})
		})
	})

	Describe("WaitUntilReady", func() {
		var (
			registryConfig biinstallmanifest.Registry
		)

		Context("When raw private key is provided", func() {
			BeforeEach(func() {
				registryConfig = biinstallmanifest.Registry{
					Port: 125,
					SSHTunnel: biinstallmanifest.SSHTunnel{
						Host:       "fake-ssh-host",
						Port:       124,
						User:       "fake-ssh-username",
						Password:   "fake-password",
						PrivateKey: "--BEGIN PRIVATE KEY-- asdf --END PRIVATE KEY--",
					},
				}
			})

			It("starts & stops the SSH tunnel", func() {
				err := instance.WaitUntilReady(registryConfig, fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bisshtunnel.Options{
					User:              "fake-ssh-username",
					PrivateKey:        "--BEGIN PRIVATE KEY-- asdf --END PRIVATE KEY--",
					Password:          "fake-password",
					Host:              "fake-ssh-host",
					Port:              124,
					LocalForwardPort:  125,
					RemoteForwardPort: 125,
				}))
				Expect(fakeSSHTunnel.Started).To(BeTrue())
			})

			It("waits for the vm", func() {
				err := instance.WaitUntilReady(registryConfig, fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebivm.WaitUntilReadyInput{
					Timeout: 10 * time.Minute,
					Delay:   500 * time.Millisecond,
				}))
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.WaitUntilReady(registryConfig, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
					{Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready"},
				}))
			})

			Context("when registry config is empty", func() {
				BeforeEach(func() {
					registryConfig = biinstallmanifest.Registry{}
				})

				It("does not start ssh tunnel", func() {
					err := instance.WaitUntilReady(registryConfig, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSSHTunnel.Started).To(BeFalse())
				})
			})

			Context("when registry config is empty", func() {
				BeforeEach(func() {
					registryConfig = biinstallmanifest.Registry{}
				})

				It("does not start ssh tunnel", func() {
					err := instance.WaitUntilReady(registryConfig, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeSSHTunnel.Started).To(BeFalse())
				})
			})

			Context("when starting SSH tunnel fails", func() {
				BeforeEach(func() {
					fakeSSHTunnel.SetStartBehavior(bosherr.Error("fake-ssh-tunnel-start-error"), nil)
				})

				It("returns an error", func() {
					err := instance.WaitUntilReady(registryConfig, fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
				})
			})

			Context("when waiting for the agent fails", func() {
				var (
					waitError = bosherr.Error("fake-wait-error")
				)
				BeforeEach(func() {
					fakeVM.WaitUntilReadyErr = waitError
				})

				It("logs start and stop events to the eventLogger", func() {
					err := instance.WaitUntilReady(registryConfig, fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

					Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
						{
							Name:  "Waiting for the agent on VM 'fake-vm-cid' to be ready",
							Error: waitError,
						},
					}))
				})
			})

			Context("when receiving SSH tunnel errors", func() {
				BeforeEach(func() {
					fakeSSHTunnel.SetStartBehavior(nil, bosherr.Error("fake-ssh-tunnel-error"))
				})

				It("logs the error", func() {
					err := instance.WaitUntilReady(registryConfig, fakeStage)
					Expect(err).NotTo(HaveOccurred())

					Eventually(logger.WarnCallCount).Should(Equal(1))
					tag, message, _ := logger.WarnArgsForCall(0)
					Expect(tag).To(Equal("instance"))
					Expect(message).To(Equal("Received SSH tunnel error: %s"))
				})
			})
		})

		Context("When the private key is provided", func() {
			BeforeEach(func() {
				registryConfig = biinstallmanifest.Registry{
					Port: 125,
					SSHTunnel: biinstallmanifest.SSHTunnel{
						Host:       "fake-ssh-host",
						Port:       124,
						User:       "fake-ssh-username",
						Password:   "fake-password",
						PrivateKey: "--BEGIN PRIVATE KEY-- asdf --END PRIVATE KEY--",
					},
				}
			})

			It("sets the SSHTunnel options", func() {
				err := instance.WaitUntilReady(registryConfig, fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bisshtunnel.Options{
					User:              "fake-ssh-username",
					PrivateKey:        "--BEGIN PRIVATE KEY-- asdf --END PRIVATE KEY--",
					Password:          "fake-password",
					Host:              "fake-ssh-host",
					Port:              124,
					LocalForwardPort:  125,
					RemoteForwardPort: 125,
				}))
			})

		})
	})

	Describe("Stop", func() {
		It("checks if the agent on the vm is responsive", func() {
			err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebivm.WaitUntilReadyInput{
				Timeout: pingTimeout,
				Delay:   pingDelay,
			}))
		})

		It("logs start and stop events", func() {
			err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
				{Name: "Running the pre-stop scripts 'fake-job-name/0'"},
				{Name: "Draining jobs on instance 'fake-job-name/0'"},
				{Name: "Stopping jobs on instance 'fake-job-name/0'"},
				{Name: "Running the post-stop scripts 'fake-job-name/0'"},
			}))
		})

		Context("when agent is responsive", func() {
			It("logs waiting for the agent event", func() {
				err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0]).To(Equal(&fakebiui.PerformCall{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
				}))
			})

			It("drains", func() {
				err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.DrainCalled).To(Equal(1))
			})

			It("can skip draining", func() {
				skipDrain = true
				err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.DrainCalled).To(Equal(0))
			})

			It("stops all jobs", func() {
				err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.StopCalled).To(Equal(1))
			})

			It("does not unmount vm disks", func() {
				firstDisk := fakebidisk.NewFakeDisk("fake-disk-1")
				secondDisk := fakebidisk.NewFakeDisk("fake-disk-2")
				fakeVM.ListDisksDisks = []bidisk.Disk{firstDisk, secondDisk}

				err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.UnmountDiskInputs).To(Equal([]fakebivm.UnmountDiskInput{}))
			})
		})

		Context("when agent fails to respond", func() {
			BeforeEach(func() {
				fakeVM.WaitUntilReadyErr = bosherr.Error("fake-wait-error")
			})

			It("logs failed event", func() {
				err := instance.Stop(pingTimeout, pingDelay, skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Waiting for the agent on VM 'fake-vm-cid'"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Agent unreachable: fake-wait-error"))
			})
		})
	})

	Describe("Start", func() {

		var (
			update bideplmanifest.Update
		)

		BeforeEach(func() {
			update = bideplmanifest.Update{
				UpdateWatchTime: bideplmanifest.WatchTime{
					Start: 0,
					End:   5478,
				},
			}
		})

		It("checks if the agent on the vm is responsive", func() {
			err := instance.Start(update, pingTimeout, pingDelay, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebivm.WaitUntilReadyInput{
				Timeout: pingTimeout,
				Delay:   pingDelay,
			}))
		})

		It("logs start and stop events", func() {
			err := instance.Start(update, pingTimeout, pingDelay, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Waiting for the agent on VM 'fake-vm-cid'"},
				{Name: "Running the pre-start scripts 'fake-job-name/0'"},
				{Name: "Starting the agent 'fake-job-name/0'"},
				{Name: "Waiting for instance 'fake-job-name/0' to be running"},
				{Name: "Running the post-start scripts 'fake-job-name/0'"},
			}))
		})

		Context("when agent is responsive", func() {
			It("logs waiting for the agent event", func() {
				err := instance.Start(update, pingTimeout, pingDelay, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0]).To(Equal(&fakebiui.PerformCall{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
				}))
			})

			It("waits until agent reports state as running", func() {
				err := instance.Start(update, pingTimeout, pingDelay, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebivm.WaitInput{
					MaxAttempts: 5,
					Delay:       1 * time.Second,
				}))
			})
		})

		Context("when agent fails to respond", func() {
			BeforeEach(func() {
				fakeVM.WaitUntilReadyErr = bosherr.Error("fake-wait-error")
			})

			It("logs failed event", func() {
				err := instance.Start(update, pingTimeout, pingDelay, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Waiting for the agent on VM 'fake-vm-cid'"))
				Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("Agent unreachable: fake-wait-error"))
			})
		})
	})
})
