package deployer_test

import (
	"errors"
	"fmt"
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmins "github.com/cloudfoundry/bosh-micro-cli/deployer/instance/fakes"
	fakeregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry/fakes"
	fakebmretry "github.com/cloudfoundry/bosh-micro-cli/deployer/retrystrategy/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/vm/fakes"
)

var _ = Describe("Deployer", func() {
	var (
		deployer                   Deployer
		fakeVMManagerFactory       *fakebmvm.FakeManagerFactory
		fakeVMManager              *fakebmvm.FakeManager
		cloud                      *fakebmcloud.FakeCloud
		deployment                 bmdepl.Deployment
		registry                   bmdepl.Registry
		fakeRegistryServer         *fakeregistry.FakeServer
		eventLogger                *fakebmlog.FakeEventLogger
		fakeSSHTunnel              *fakebmsshtunnel.FakeTunnel
		fakeSSHTunnelFactory       *fakebmsshtunnel.FakeFactory
		fakeInstance               *fakebmins.FakeInstance
		sshTunnelConfig            bmdepl.SSHTunnel
		fakeAgentPingRetryStrategy *fakebmretry.FakeRetryStrategy

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
		fakeSSHTunnelFactory = fakebmsshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebmsshtunnel.NewFakeTunnel()
		fakeSSHTunnel.SetStartBehavior(nil, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel
		fakeInstance = fakebmins.NewFakeInstance()
		instanceFactory := fakebmins.NewFakeInstanceFactory()
		instanceFactory.CreateInstance = fakeInstance

		logger := boshlog.NewLogger(boshlog.LevelNone)
		eventLogger = fakebmlog.NewFakeEventLogger()
		deployer = NewDeployer(
			fakeVMManagerFactory,
			fakeSSHTunnelFactory,
			fakeRegistryServer,
			instanceFactory,
			eventLogger,
			logger,
		)
		fakeAgentPingRetryStrategy = fakebmretry.NewFakeRetryStrategy()

		applySpec = bmstemcell.ApplySpec{
			Job: bmstemcell.Job{
				Name: "fake-job-name",
			},
		}
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
		Expect(fakeVMManager.CreateVMInput).To(Equal(
			fakebmvm.CreateVMInput{
				StemcellCID: "fake-stemcell-cid",
				Deployment:  deployment,
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

	It("waits for the instance", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeInstance.WaitToBeReadyInputs).To(ContainElement(fakebmins.WaitInput{
			MaxAttempts: 300,
			Delay:       500 * time.Millisecond,
		}))
	})

	It("updates the instance", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeInstance.ApplyInputs).To(ContainElement(fakebmins.ApplyInput{
			StemcellApplySpec: applySpec,
			Deployment:        deployment,
		}))
	})

	It("starts the agent", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeInstance.StartCalled).To(BeTrue())
	})

	It("waits until agent reports state as running", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeInstance.WaitToBeRunningInputs).To(ContainElement(fakebmins.WaitInput{
			MaxAttempts: 5,
			Delay:       1 * time.Second,
		}))
	})

	It("logs start and stop events to the eventLogger", func() {
		err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Waiting for the agent"),
			Index: 2,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Waiting for the agent"),
			Index: 2,
			State: bmeventlog.Finished,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Applying micro BOSH spec"),
			Index: 3,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Applying micro BOSH spec"),
			Index: 3,
			State: bmeventlog.Finished,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Starting agent services"),
			Index: 4,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Starting agent services"),
			Index: 4,
			State: bmeventlog.Finished,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Waiting for the director"),
			Index: 5,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 5,
			Task:  fmt.Sprintf("Waiting for the director"),
			Index: 5,
			State: bmeventlog.Finished,
		}))

		Expect(eventLogger.LoggedEvents).To(HaveLen(8))
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
			fakeInstance.WaitToBeReadyErr = errors.New("fake-wait-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 5,
				Task:  fmt.Sprintf("Waiting for the agent"),
				Index: 2,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   5,
				Task:    fmt.Sprintf("Waiting for the agent"),
				Index:   2,
				State:   bmeventlog.Failed,
				Message: "fake-wait-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(2))
		})
	})

	Context("when updating instance fails", func() {
		BeforeEach(func() {
			fakeInstance.ApplyErr = errors.New("fake-apply-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-apply-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 5,
				Task:  fmt.Sprintf("Applying micro BOSH spec"),
				Index: 3,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   5,
				Task:    fmt.Sprintf("Applying micro BOSH spec"),
				Index:   3,
				State:   bmeventlog.Failed,
				Message: "fake-apply-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(4))
		})
	})

	Context("when starting agent services fails", func() {
		BeforeEach(func() {
			fakeInstance.StartErr = errors.New("fake-start-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-start-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 5,
				Task:  fmt.Sprintf("Starting agent services"),
				Index: 4,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   5,
				Task:    fmt.Sprintf("Starting agent services"),
				Index:   4,
				State:   bmeventlog.Failed,
				Message: "fake-start-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(6))
		})
	})

	Context("when creating VM fails", func() {
		It("returns an error", func() {
			createVMError := errors.New("fake-create-vm-error")
			fakeVMManager.SetCreateVMBehavior("fake-stemcell-cid", createVMError)
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
		})
	})

	Context("when waiting for running state fails", func() {
		BeforeEach(func() {
			fakeInstance.WaitToBeRunningErr = errors.New("fake-wait-running-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, applySpec, registry, sshTunnelConfig, "fake-mbus-url", "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 5,
				Task:  fmt.Sprintf("Waiting for the director"),
				Index: 5,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   5,
				Task:    fmt.Sprintf("Waiting for the director"),
				Index:   5,
				State:   bmeventlog.Failed,
				Message: "fake-wait-running-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(8))
		})
	})
})
