package deployer_test

import (
	"errors"
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk/fakes"
	fakeregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry/fakes"
	fakebmretry "github.com/cloudfoundry/bosh-micro-cli/deployer/retrystrategy/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Deployer", func() {
	var (
		deployer                   Deployer
		fakeVMManagerFactory       *fakebmvm.FakeManagerFactory
		fakeVMManager              *fakebmvm.FakeManager
		fakeDiskManager            *fakebmdisk.FakeManager
		cloud                      *fakebmcloud.FakeCloud
		deployment                 bmdepl.Deployment
		registry                   bmdepl.Registry
		fakeRegistryServer         *fakeregistry.FakeServer
		eventLogger                *fakebmlog.FakeEventLogger
		fakeStage                  *fakebmlog.FakeStage
		fakeSSHTunnel              *fakebmsshtunnel.FakeTunnel
		fakeSSHTunnelFactory       *fakebmsshtunnel.FakeFactory
		sshTunnelConfig            bmdepl.SSHTunnel
		fakeAgentPingRetryStrategy *fakebmretry.FakeRetryStrategy
		fakeVM                     *fakebmvm.FakeVM

		applySpec bmstemcell.ApplySpec
	)

	BeforeEach(func() {
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
		registry = bmdepl.Registry{
			Username: "fake-username",
			Password: "fake-password",
			Host:     "fake-host",
			Port:     123,
		}
		sshTunnelConfig = bmdepl.SSHTunnel{
			User:       "fake-ssh-username",
			PrivateKey: "fake-private-key-path",
			Password:   "fake-password",
			Host:       "fake-ssh-host",
			Port:       124,
		}

		cloud = fakebmcloud.NewFakeCloud()
		fakeRegistryServer = fakeregistry.NewFakeServer()
		fakeVMManagerFactory = fakebmvm.NewFakeManagerFactory()
		fakeVMManager = fakebmvm.NewFakeManager()
		fakeVMManagerFactory.SetNewManagerBehavior(cloud, fakeVMManager)

		fakeDiskManagerFactory := fakebmdisk.NewFakeManagerFactory()
		fakeDiskManager = fakebmdisk.NewFakeManager()
		fakeDiskManager.CreateDisk = bmdisk.NewDisk("fake-disk-cid")
		fakeDiskManagerFactory.NewManagerManager = fakeDiskManager

		fakeSSHTunnelFactory = fakebmsshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebmsshtunnel.NewFakeTunnel()
		fakeSSHTunnel.SetStartBehavior(nil, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel

		logger := boshlog.NewLogger(boshlog.LevelNone)
		eventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		eventLogger.SetNewStageBehavior(fakeStage)
		deployer = NewDeployer(
			fakeVMManagerFactory,
			fakeDiskManagerFactory,
			fakeSSHTunnelFactory,
			fakeRegistryServer,
			eventLogger,
			logger,
		)
		fakeAgentPingRetryStrategy = fakebmretry.NewFakeRetryStrategy()

		applySpec = bmstemcell.ApplySpec{
			Job: bmstemcell.Job{
				Name: "fake-job-name",
			},
		}

		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")
		fakeVMManager.CreateVM = fakeVM
	})

	It("adds a new event logger stage", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).ToNot(HaveOccurred())

		Expect(eventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
			{
				Name: "deploying",
			},
		}))

		Expect(fakeStage.Started).To(BeTrue())
		Expect(fakeStage.Finished).To(BeTrue())
	})

	It("starts the registry", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeRegistryServer.StartInput).To(Equal(fakeregistry.StartInput{
			Username: "fake-username",
			Password: "fake-password",
			Host:     "fake-host",
			Port:     123,
		}))
		Expect(fakeRegistryServer.ReceivedActions).To(Equal([]string{"Start", "Stop"}))
	})

	It("creates a VM", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVMManager.CreateInput).To(Equal(
			fakebmvm.CreateInput{
				StemcellCID: "fake-stemcell-cid",
				Deployment:  deployment,
				MbusURL:     "fake-mbus-url",
			},
		))
	})

	It("starts the SSH tunnel", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeSSHTunnel.Started).To(BeTrue())
		Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bmsshtunnel.Options{
			User:              "fake-ssh-username",
			PrivateKey:        "fake-private-key-path",
			Password:          "fake-password",
			Host:              "fake-ssh-host",
			Port:              124,
			LocalForwardPort:  123,
			RemoteForwardPort: 123,
		}))
	})

	It("waits for the vm", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVM.WaitToBeReadyInputs).To(ContainElement(fakebmvm.WaitInput{
			MaxAttempts: 300,
			Delay:       500 * time.Millisecond,
		}))
	})

	It("updates the vm", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.ApplyInputs).To(ContainElement(fakebmvm.ApplyInput{
			StemcellApplySpec: applySpec,
			Deployment:        deployment,
		}))
	})

	It("starts the agent", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.StartCalled).To(BeTrue())
	})

	It("waits until agent reports state as running", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebmvm.WaitInput{
			MaxAttempts: 5,
			Delay:       1 * time.Second,
		}))
	})

	Context("when the deployment has a disk pool", func() {
		var diskPool bmdepl.DiskPool

		BeforeEach(func() {
			diskPool = bmdepl.DiskPool{
				Name: "fake-persistent-disk-pool-name",
				Size: 1024,
				RawCloudProperties: map[interface{}]interface{}{
					"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
				},
			}
			deployment.DiskPools = []bmdepl.DiskPool{diskPool}
			deployment.Jobs[0].PersistentDiskPool = "fake-persistent-disk-pool-name"
		})

		It("creates a persistent disk", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDiskManager.CreateInputs).To(Equal([]fakebmdisk.CreateInput{
				{
					DiskPool:   diskPool,
					InstanceID: "fake-vm-cid",
				},
			}))
		})

		It("attaches the persistent disk", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.AttachDiskInputs).To(Equal([]fakebmvm.AttachDiskInput{
				{
					Disk: bmdisk.NewDisk("fake-disk-cid"),
				},
			}))
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Creating disk",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Attaching disk 'fake-disk-cid' to VM 'fake-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when creating the persistent disk fails", func() {
			BeforeEach(func() {
				fakeDiskManager.CreateErr = errors.New("fake-create-disk-error")
			})

			It("return an error", func() {
				err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-disk-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Creating disk",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-create-disk-error",
				}))
			})
		})

		Context("when attaching the persistent disk fails", func() {
			BeforeEach(func() {
				fakeVM.AttachDiskErr = errors.New("fake-attach-disk-error")
			})

			It("return an error", func() {
				err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-attach-disk-error"))
			})

			It("logs start and failed events to the eventLogger", func() {
				err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Attaching disk 'fake-disk-cid' to VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-attach-disk-error",
				}))
			})
		})
	})

	Context("when the deployment has an invalid disk pool specification", func() {
		BeforeEach(func() {
			deployment.Jobs[0].PersistentDiskPool = "fake-persistent-disk-pool-name"
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the deployment's first job does not contain a non-zero persistent disk", func() {
		BeforeEach(func() {
			deployment.Jobs[0].PersistentDisk = 0
		})

		It("does not create a persistent disk", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDiskManager.CreateInputs).To(BeEmpty())
		})
	})

	It("logs start and stop events to the eventLogger", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

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
		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Starting 'fake-job-name'",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Waiting for 'fake-job-name'",
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
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
		})
	})

	Context("when starting registry fails", func() {
		BeforeEach(func() {
			fakeRegistryServer.SetStartBehavior(errors.New("fake-registry-start-error"), nil)
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-start-error"))
		})
	})

	Context("when waiting for the agent fails", func() {
		BeforeEach(func() {
			fakeVM.WaitToBeReadyErr = errors.New("fake-wait-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for the agent on VM 'fake-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-wait-error",
			}))
		})
	})

	Context("when updating instance fails", func() {
		BeforeEach(func() {
			fakeVM.ApplyErr = errors.New("fake-apply-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-apply-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Starting 'fake-job-name'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-apply-error",
			}))
		})
	})

	Context("when starting agent services fails", func() {
		BeforeEach(func() {
			fakeVM.StartErr = errors.New("fake-start-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-start-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Starting 'fake-job-name'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-start-error",
			}))
		})
	})

	Context("when creating VM fails", func() {
		BeforeEach(func() {
			fakeVMManager.CreateErr = errors.New("fake-create-vm-error")
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Creating VM from stemcell 'fake-stemcell-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-create-vm-error",
			}))
		})
	})

	Context("when waiting for running state fails", func() {
		BeforeEach(func() {
			fakeVM.WaitToBeRunningErr = errors.New("fake-wait-running-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for 'fake-job-name'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-wait-running-error",
			}))
		})
	})
})
