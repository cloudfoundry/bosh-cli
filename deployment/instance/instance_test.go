package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"

	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec/fakes"
	fakebmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Instance", func() {
	var (
		fakeVMManager              *fakebmvm.FakeManager
		fakeVM                     *fakebmvm.FakeVM
		fakeSSHTunnelFactory       *fakebmsshtunnel.FakeFactory
		fakeSSHTunnel              *fakebmsshtunnel.FakeTunnel
		fakeTemplatesSpecGenerator *fakebmas.FakeTemplatesSpecGenerator
		fakeStage                  *fakebmlog.FakeStage

		blobstoreURL = "https://fake-blobstore-url"

		instance Instance

		pingTimeout = 1 * time.Second
		pingDelay   = 500 * time.Millisecond
	)

	BeforeEach(func() {
		fakeVMManager = fakebmvm.NewFakeManager()
		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")

		fakeSSHTunnelFactory = fakebmsshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebmsshtunnel.NewFakeTunnel()
		fakeSSHTunnel.SetStartBehavior(nil, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel

		fakeTemplatesSpecGenerator = fakebmas.NewFakeTemplatesSpecGenerator()

		logger := boshlog.NewLogger(boshlog.LevelNone)

		instance = NewInstance(
			"fake-job-name",
			0,
			fakeVM,
			fakeVMManager,
			fakeSSHTunnelFactory,
			fakeTemplatesSpecGenerator,
			blobstoreURL,
			logger,
		)

		fakeStage = fakebmlog.NewFakeStage()
	})

	Describe("Delete", func() {
		It("checks if the agent on the vm is responsive", func() {
			err := instance.Delete(pingTimeout, pingDelay, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebmvm.WaitUntilReadyInput{
				Timeout: pingTimeout,
				Delay:   pingDelay,
			}))
		})

		It("deletes existing vm", func() {
			err := instance.Delete(pingTimeout, pingDelay, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.DeleteCalled).To(Equal(1))
		})

		It("logs start and stop events", func() {
			err := instance.Delete(pingTimeout, pingDelay, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(Equal([]*fakebmlog.FakeStep{
				{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				{
					Name: "Stopping jobs on instance 'fake-job-name/0'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				{
					Name: "Deleting VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
			}))
		})

		Context("when agent is responsive", func() {
			It("logs waiting for the agent event", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				}))
			})

			It("stops vm", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.StopCalled).To(Equal(1))
			})

			It("unmounts vm disks", func() {
				firstDisk := fakebmdisk.NewFakeDisk("fake-disk-1")
				secondDisk := fakebmdisk.NewFakeDisk("fake-disk-2")
				fakeVM.ListDisksDisks = []bmdisk.Disk{firstDisk, secondDisk}

				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVM.UnmountDiskInputs).To(Equal([]fakebmvm.UnmountDiskInput{
					{Disk: firstDisk},
					{Disk: secondDisk},
				}))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Unmounting disk 'fake-disk-1'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				}))
				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Unmounting disk 'fake-disk-2'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				}))
			})

			Context("when stopping vm fails", func() {
				BeforeEach(func() {
					fakeVM.StopErr = errors.New("fake-stop-error")
				})

				It("returns an error", func() {
					err := instance.Delete(pingTimeout, pingDelay, fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-stop-error"))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Stopping jobs on instance 'fake-job-name/0'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Failed,
						},
						FailMessage: "fake-stop-error",
					}))
				})
			})

			Context("when unmounting disk fails", func() {
				BeforeEach(func() {
					fakeVM.ListDisksDisks = []bmdisk.Disk{fakebmdisk.NewFakeDisk("fake-disk")}
					fakeVM.UnmountDiskErr = errors.New("fake-unmount-error")
				})

				It("returns an error", func() {
					err := instance.Delete(pingTimeout, pingDelay, fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-unmount-error"))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Unmounting disk 'fake-disk'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Failed,
						},
						FailMessage: "Unmounting disk 'fake-disk' from VM 'fake-vm-cid': fake-unmount-error",
					}))
				})
			})
		})

		Context("when agent fails to respond", func() {
			BeforeEach(func() {
				fakeVM.WaitUntilReadyErr = errors.New("fake-wait-error")
			})

			It("logs failed event", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Agent unreachable: fake-wait-error",
				}))
			})
		})

		Context("when deleting VM fails", func() {
			BeforeEach(func() {
				fakeVM.DeleteErr = errors.New("fake-delete-error")
			})

			It("returns an error", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Deleting VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-delete-error",
				}))
			})
		})

		Context("when VM does not exist (deleted manually)", func() {
			BeforeEach(func() {
				fakeVM.ExistsFound = false
				fakeVM.DeleteErr = bmcloud.NewCPIError("delete_vm", bmcloud.CmdError{
					Type:    bmcloud.VMNotFoundError,
					Message: "fake-vm-not-found-message",
				})
			})

			It("deletes existing vm", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeVM.DeleteCalled).To(Equal(1))
			})

			It("does not contact the agent", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeVM.WaitUntilReadyInputs).To(HaveLen(0))
				Expect(fakeVM.StopCalled).To(Equal(0))
				Expect(fakeVM.UnmountDiskInputs).To(HaveLen(0))
			})

			It("logs vm delete as skipped", func() {
				err := instance.Delete(pingTimeout, pingDelay, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStage.Steps).To(Equal([]*fakebmlog.FakeStep{
					{
						Name: "Deleting VM 'fake-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Skipped,
						},
						SkipMessage: "CPI 'delete_vm' method responded with error: CmdError{\"type\":\"Bosh::Cloud::VMNotFound\",\"message\":\"fake-vm-not-found-message\",\"ok_to_retry\":false}",
					},
				}))
			})
		})
	})

	Describe("UpdateJobs", func() {
		var (
			jobBlobs           []bmstemcell.Blob
			applySpec          bmstemcell.ApplySpec
			deploymentJob      bmdeplmanifest.Job
			deploymentManifest bmdeplmanifest.Manifest
		)

		BeforeEach(func() {
			jobBlobs = []bmstemcell.Blob{
				{
					Name:        "first-job-name",
					Version:     "first-job-version",
					SHA1:        "first-job-sha1",
					BlobstoreID: "first-job-blobstore-id",
				},
				{
					Name:        "second-job-name",
					Version:     "second-job-version",
					SHA1:        "second-job-sha1",
					BlobstoreID: "second-job-blobstore-id",
				},
				{
					Name:        "third-job-name",
					Version:     "third-job-version",
					SHA1:        "third-job-sha1",
					BlobstoreID: "third-job-blobstore-id",
				},
			}

			//TODO: compile packages too
			applySpec = bmstemcell.ApplySpec{
				//				Packages: map[string]bmstemcell.Blob{
				//					"first-package-name": bmstemcell.Blob{
				//						Name:        "first-package-name",
				//						Version:     "first-package-version",
				//						SHA1:        "first-package-sha1",
				//						BlobstoreID: "first-package-blobstore-id",
				//					},
				//					"second-package-name": bmstemcell.Blob{
				//						Name:        "second-package-name",
				//						Version:     "second-package-version",
				//						SHA1:        "second-package-sha1",
				//						BlobstoreID: "second-package-blobstore-id",
				//					},
				//				},
				Job: bmstemcell.Job{
					Name:      "fake-job-name",
					Templates: jobBlobs,
				},
			}

			deploymentJob = bmdeplmanifest.Job{
				Name: "fake-job-name",
				Templates: []bmdeplmanifest.ReleaseJobRef{
					{Name: "first-job-name"},
					{Name: "third-job-name"},
				},
				PersistentDiskPool: "fake-persistent-disk-pool-name",
				RawProperties: map[interface{}]interface{}{
					"fake-property-key": "fake-property-value",
				},
				Networks: []bmdeplmanifest.JobNetwork{
					{
						Name:      "fake-network-name",
						StaticIPs: []string{"fake-network-ip"},
					},
				},
			}

			deploymentManifest = bmdeplmanifest.Manifest{
				Name: "fake-deployment-name",
				Networks: []bmdeplmanifest.Network{
					{
						Name:               "fake-network-name",
						Type:               bmdeplmanifest.NetworkType("fake-network-type"),
						RawCloudProperties: map[interface{}]interface{}{},
						//						IP: "",
						//						Netmask: "",
						//						Gateway: "",
						//						DNS: []string{},
					},
				},
				Update: bmdeplmanifest.Update{
					UpdateWatchTime: bmdeplmanifest.WatchTime{
						Start: 0,
						End:   5478,
					},
				},
				Jobs: []bmdeplmanifest.Job{
					deploymentJob,
				},
			}
		})

		JustBeforeEach(func() {
			fakeTemplatesSpecGenerator.CreateTemplatesSpec = bmas.TemplatesSpec{
				BlobID:            "fake-blob-id",
				ArchiveSha1:       "fake-archive-sha1",
				ConfigurationHash: "fake-configuration-hash",
			}
		})

		It("renders jobs templates (multiple jobs with multiple templates each)", func() {
			err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeTemplatesSpecGenerator.CreateTemplatesSpecInputs).To(ContainElement(fakebmas.CreateTemplatesSpecInput{
				DeploymentJob:  deploymentJob,
				JobBlobs:       jobBlobs,
				DeploymentName: "fake-deployment-name",
				Properties: map[string]interface{}{
					"fake-property-key": "fake-property-value",
				},
				MbusURL: blobstoreURL,
			}))
		})

		It("tells agent to stop jobs, apply a new spec (with new rendered jobs templates), and start jobs", func() {
			err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.StartCalled).To(Equal(1))
			//TODO: test compiled packages too
			expectedApplySpec := bmas.ApplySpec{
				Deployment: "fake-deployment-name",
				Index:      0,
				Packages:   map[string]bmas.Blob{},
				Networks: map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"type":             "fake-network-type",
						"ip":               "fake-network-ip",
						"cloud_properties": map[string]interface{}{},
					},
				},
				Job: bmas.Job{
					Name: "fake-job-name",
					Templates: []bmas.Blob{
						{
							Name:        "first-job-name",
							Version:     "first-job-version",
							SHA1:        "first-job-sha1",
							BlobstoreID: "first-job-blobstore-id",
						},
						// TODO: remove second-job, because the deployment job doesn't reference it
						{
							Name:        "second-job-name",
							Version:     "second-job-version",
							SHA1:        "second-job-sha1",
							BlobstoreID: "second-job-blobstore-id",
						},
						{
							Name:        "third-job-name",
							Version:     "third-job-version",
							SHA1:        "third-job-sha1",
							BlobstoreID: "third-job-blobstore-id",
						},
					},
				},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{
					BlobstoreID: "fake-blob-id",
					SHA1:        "fake-archive-sha1",
				},
				ConfigurationHash: "fake-configuration-hash",
			}
			Expect(fakeVM.ApplyInputs).To(Equal([]fakebmvm.ApplyInput{
				{ApplySpec: expectedApplySpec},
			}))
			Expect(fakeVM.StopCalled).To(Equal(1))
		})

		It("waits until agent reports state as running", func() {
			err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebmvm.WaitInput{
				MaxAttempts: 5,
				Delay:       1 * time.Second,
			}))
		})

		It("logs start and stop events to the eventLogger", func() {
			err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Updating instance 'fake-job-name/0'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for instance 'fake-job-name/0' to be running",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when rendering jobs templates fails", func() {
			BeforeEach(func() {
				fakeTemplatesSpecGenerator.CreateErr = errors.New("fake-template-err")
			})

			It("returns an error", func() {
				err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-err"))
			})
		})

		Context("when stopping vm fails", func() {
			BeforeEach(func() {
				fakeVM.StopErr = errors.New("fake-stop-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stop-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Updating instance 'fake-job-name/0'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Stopping the agent: fake-stop-error",
				}))
			})
		})

		Context("when applying a new vm state fails", func() {
			BeforeEach(func() {
				fakeVM.ApplyErr = errors.New("fake-apply-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-apply-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Updating instance 'fake-job-name/0'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Applying the agent state: fake-apply-error",
				}))
			})
		})

		Context("when starting vm fails", func() {
			BeforeEach(func() {
				fakeVM.StartErr = errors.New("fake-start-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-start-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Updating instance 'fake-job-name/0'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Starting the agent: fake-start-error",
				}))
			})
		})

		Context("when waiting for running state fails", func() {
			BeforeEach(func() {
				fakeVM.WaitToBeRunningErr = errors.New("fake-wait-running-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.UpdateJobs(deploymentManifest, applySpec, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Waiting for instance 'fake-job-name/0' to be running",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-wait-running-error",
				}))
			})
		})
	})

	Describe("WaitUntilReady", func() {
		var (
			registryConfig  bminstallmanifest.Registry
			sshTunnelConfig bminstallmanifest.SSHTunnel
		)

		BeforeEach(func() {
			registryConfig = bminstallmanifest.Registry{
				Port: 125,
			}
			sshTunnelConfig = bminstallmanifest.SSHTunnel{
				Host:       "fake-ssh-host",
				Port:       124,
				User:       "fake-ssh-username",
				Password:   "fake-password",
				PrivateKey: "fake-private-key-path",
			}
		})

		It("starts & stops the SSH tunnel", func() {
			err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bmsshtunnel.Options{
				User:              "fake-ssh-username",
				PrivateKey:        "fake-private-key-path",
				Password:          "fake-password",
				Host:              "fake-ssh-host",
				Port:              124,
				LocalForwardPort:  125,
				RemoteForwardPort: 125,
			}))
			Expect(fakeSSHTunnel.Started).To(BeTrue())
			Expect(fakeSSHTunnel.Stopped).To(BeTrue())
		})

		It("waits for the vm", func() {
			err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebmvm.WaitUntilReadyInput{
				Timeout: 10 * time.Minute,
				Delay:   500 * time.Millisecond,
			}))
		})

		It("logs start and stop events to the eventLogger", func() {
			err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when ssh tunnel config is empty", func() {
			BeforeEach(func() {
				sshTunnelConfig = bminstallmanifest.SSHTunnel{}
			})

			It("does not start ssh tunnel", func() {
				err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeSSHTunnel.Started).To(BeFalse())
			})
		})

		Context("when registry config is empty", func() {
			BeforeEach(func() {
				registryConfig = bminstallmanifest.Registry{}
			})

			It("does not start ssh tunnel", func() {
				err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeSSHTunnel.Started).To(BeFalse())
			})
		})

		Context("when starting SSH tunnel fails", func() {
			BeforeEach(func() {
				fakeSSHTunnel.SetStartBehavior(errors.New("fake-ssh-tunnel-start-error"), nil)
			})

			It("returns an error", func() {
				err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
			})
		})

		Context("when waiting for the agent fails", func() {
			BeforeEach(func() {
				fakeVM.WaitUntilReadyErr = errors.New("fake-wait-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-wait-error",
				}))
			})
		})
	})
})
