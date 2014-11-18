package vm_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
)

var _ = Describe("VmDeployer", func() {
	var (
		fakeVMManagerFactory *fakebmvm.FakeManagerFactory
		fakeVMManager        *fakebmvm.FakeManager
		fakeSSHTunnel        *fakebmsshtunnel.FakeTunnel
		fakeSSHTunnelFactory *fakebmsshtunnel.FakeFactory
		cloud                *fakebmcloud.FakeCloud
		deployment           bmdepl.Deployment
		stemcell             bmstemcell.CloudStemcell
		sshTunnelOptions     bmsshtunnel.Options
		fakeStage            *fakebmlog.FakeStage
		fakeVM               *fakebmvm.FakeVM
		vmDeployer           Deployer
	)

	BeforeEach(func() {
		fakeVMManagerFactory = fakebmvm.NewFakeManagerFactory()
		fakeVMManager = fakebmvm.NewFakeManager()
		fakeVMManagerFactory.SetNewManagerBehavior(cloud, "fake-mbus-url", fakeVMManager)
		fakeSSHTunnelFactory = fakebmsshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebmsshtunnel.NewFakeTunnel()
		fakeSSHTunnel.SetStartBehavior(nil, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel

		logger := boshlog.NewLogger(boshlog.LevelNone)

		vmDeployer = NewDeployer(fakeVMManagerFactory, fakeSSHTunnelFactory, logger)

		deployment = bmdepl.Deployment{
			Update: bmdepl.Update{
				UpdateWatchTime: bmdepl.WatchTime{
					Start: 0,
					End:   5478,
				},
			},
			Jobs: []bmdepl.Job{
				{
					Name: "fake-job-name",
				},
			},
		}

		sshTunnelOptions = bmsshtunnel.Options{
			Host:              "fake-ssh-host",
			Port:              124,
			User:              "fake-ssh-username",
			Password:          "fake-password",
			PrivateKey:        "fake-private-key-path",
			LocalForwardPort:  125,
			RemoteForwardPort: 126,
		}

		stemcell = bmstemcell.CloudStemcell{
			CID: "fake-stemcell-cid",
		}

		fakeStage = fakebmlog.NewFakeStage()

		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")
		fakeVMManager.CreateVM = fakeVM
	})

	Describe("Deploy", func() {
		Context("when vm is already deployed", func() {
			var existingVM *fakebmvm.FakeVM

			BeforeEach(func() {
				existingVM = fakebmvm.NewFakeVM("existing-vm-cid")
				fakeVMManager.SetFindCurrentBehavior(existingVM, true, nil)
			})

			It("checks if the agent on the vm is responsive", func() {
				vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(vm).To(Equal(fakeVM))

				Expect(existingVM.WaitToBeReadyInputs).To(ContainElement(fakebmvm.WaitToBeReadyInput{
					Timeout: 10 * time.Second,
					Delay:   500 * time.Millisecond,
				}))
			})

			It("deletes existing vm", func() {
				vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(vm).To(Equal(fakeVM))

				Expect(existingVM.DeleteCalled).To(Equal(1))
			})

			It("logs start and stop events to the eventLogger", func() {
				vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(vm).To(Equal(fakeVM))

				Expect(fakeStage.Steps).To(Equal([]*fakebmlog.FakeStep{
					{
						Name: "Waiting for the agent on VM 'existing-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
					{
						Name: "Stopping 'fake-job-name'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
					{
						Name: "Deleting VM 'existing-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
					{
						Name: "Creating VM from stemcell 'fake-stemcell-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
					{
						Name: "Waiting for the agent on VM 'fake-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					},
				}))
			})

			Context("when agent is responsive", func() {
				It("logs waiting for the agent event", func() {
					_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Waiting for the agent on VM 'existing-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					}))
				})

				It("stops vm", func() {
					_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(existingVM.StopCalled).To(Equal(1))
				})

				It("unmounts vm disks", func() {
					existingVM.ListDisksDisks = []bmdisk.Disk{bmdisk.NewDisk("fake-disk-1"), bmdisk.NewDisk("fake-disk-2")}

					_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(existingVM.UnmountDiskInputs).To(Equal([]fakebmvm.UnmountDiskInput{
						{
							Disk: bmdisk.NewDisk("fake-disk-1"),
						},
						{
							Disk: bmdisk.NewDisk("fake-disk-2"),
						},
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
						existingVM.StopErr = errors.New("fake-stop-error")
					})

					It("returns an error", func() {
						_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-stop-error"))

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Stopping 'fake-job-name'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Failed,
							},
							FailMessage: "Stopping VM: fake-stop-error",
						}))
					})
				})

				Context("when unmounting disk fails", func() {
					BeforeEach(func() {
						existingVM.ListDisksDisks = []bmdisk.Disk{bmdisk.NewDisk("fake-disk")}
						existingVM.UnmountDiskErr = errors.New("fake-unmount-error")
					})

					It("returns an error", func() {
						_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-unmount-error"))

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Unmounting disk 'fake-disk'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Failed,
							},
							FailMessage: "Unmounting disk 'fake-disk' from VM 'existing-vm-cid': fake-unmount-error",
						}))
					})
				})
			})

			Context("when agent fails to respond", func() {
				BeforeEach(func() {
					existingVM.WaitToBeReadyErr = errors.New("fake-wait-error")
				})

				It("logs failed event", func() {
					_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Waiting for the agent on VM 'existing-vm-cid'",
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
					existingVM.DeleteErr = errors.New("fake-delete-error")
				})

				It("returns an error", func() {
					_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-delete-error"))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Deleting VM 'existing-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Failed,
						},
						FailMessage: "Deleting VM: fake-delete-error",
					}))
				})
			})
		})

		It("creates a VM", func() {
			vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(vm).To(Equal(fakeVM))
			Expect(fakeVMManager.CreateInput).To(Equal(fakebmvm.CreateInput{
				Stemcell:   stemcell,
				Deployment: deployment,
			}))
		})

		It("starts the SSH tunnel", func() {
			vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(vm).To(Equal(fakeVM))
			Expect(fakeSSHTunnel.Started).To(BeTrue())
			Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bmsshtunnel.Options{
				User:              "fake-ssh-username",
				PrivateKey:        "fake-private-key-path",
				Password:          "fake-password",
				Host:              "fake-ssh-host",
				Port:              124,
				LocalForwardPort:  125,
				RemoteForwardPort: 126,
			}))
		})

		It("waits for the vm", func() {
			vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(vm).To(Equal(fakeVM))
			Expect(fakeVM.WaitToBeReadyInputs).To(ContainElement(fakebmvm.WaitToBeReadyInput{
				Timeout: 10 * time.Minute,
				Delay:   500 * time.Millisecond,
			}))
		})

		It("logs start and stop events to the eventLogger", func() {
			vm, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(vm).To(Equal(fakeVM))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Creating VM from stemcell 'fake-stemcell-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for the agent on VM 'fake-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when starting SSH tunnel fails", func() {
			BeforeEach(func() {
				fakeSSHTunnel.SetStartBehavior(errors.New("fake-ssh-tunnel-start-error"), nil)
			})

			It("returns an error", func() {
				_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
			})
		})

		Context("when waiting for the agent fails", func() {
			BeforeEach(func() {
				fakeVM.WaitToBeReadyErr = errors.New("fake-wait-error")
			})

			It("logs start and stop events to the eventLogger", func() {
				_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Waiting for the agent on VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Waiting for the vm to be ready: fake-wait-error",
				}))
			})
		})

		Context("when creating VM fails", func() {
			BeforeEach(func() {
				fakeVMManager.CreateErr = errors.New("fake-create-vm-error")
			})

			It("returns an error", func() {
				_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				_, err := vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, "fake-mbus-url", fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Creating VM from stemcell 'fake-stemcell-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Creating VM: fake-create-vm-error",
				}))
			})
		})
	})
})
