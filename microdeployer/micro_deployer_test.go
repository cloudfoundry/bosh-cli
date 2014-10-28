package microdeployer_test

import (
	"errors"
	"fmt"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/sshtunnel"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging/fakes"
	fakebminsup "github.com/cloudfoundry/bosh-micro-cli/instanceupdater/fakes"
	fakeregistry "github.com/cloudfoundry/bosh-micro-cli/registry/fakes"
	fakebmretry "github.com/cloudfoundry/bosh-micro-cli/retrystrategy/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/sshtunnel/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/vm/fakes"
)

var _ = Describe("MicroDeployer", func() {
	var (
		microDeployer              Deployer
		fakeVMManagerFactory       *fakebmvm.FakeManagerFactory
		fakeVMManager              *fakebmvm.FakeManager
		cloud                      *fakebmcloud.FakeCloud
		deployment                 bmdepl.Deployment
		registry                   bmdepl.Registry
		fakeRegistryServer         *fakeregistry.FakeServer
		eventLogger                *fakebmlog.FakeEventLogger
		fakeSSHTunnel              *fakebmsshtunnel.FakeTunnel
		fakeSSHTunnelFactory       *fakebmsshtunnel.FakeFactory
		fakeInstanceUpdater        *fakebminsup.FakeInstanceUpdater
		sshTunnelConfig            bmdepl.SSHTunnel
		fakeAgentPingRetryStrategy *fakebmretry.FakeRetryStrategy
	)

	BeforeEach(func() {
		deployment = bmdepl.Deployment{}
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
		fakeInstanceUpdater = fakebminsup.NewFakeInstanceUpdater()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		eventLogger = fakebmlog.NewFakeEventLogger()
		microDeployer = NewMicroDeployer(
			fakeVMManagerFactory,
			fakeSSHTunnelFactory,
			fakeRegistryServer,
			eventLogger,
			logger,
		)
		fakeAgentPingRetryStrategy = fakebmretry.NewFakeRetryStrategy()
		fakeInstanceUpdater = fakebminsup.NewFakeInstanceUpdater()
	})

	It("starts the registry", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
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
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVMManager.CreateVMInput).To(Equal(
			fakebmvm.CreateVMInput{
				StemcellCID: "fake-stemcell-cid",
				Deployment:  deployment,
			},
		))
	})

	It("starts the SSH tunnel", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
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

	It("waits for the agent", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeAgentPingRetryStrategy.TryCalled).To(BeTrue())
	})

	It("updates the instance", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeInstanceUpdater.UpdateCalled).To(BeTrue())
	})

	It("starts the agent", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeInstanceUpdater.StartCalled).To(BeTrue())
	})

	It("logs start and stop events to the eventLogger", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
		Expect(err).NotTo(HaveOccurred())

		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 4,
			Task:  fmt.Sprintf("Waiting for the agent"),
			Index: 2,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 4,
			Task:  fmt.Sprintf("Waiting for the agent"),
			Index: 2,
			State: bmeventlog.Finished,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 4,
			Task:  fmt.Sprintf("Applying micro BOSH spec"),
			Index: 3,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 4,
			Task:  fmt.Sprintf("Applying micro BOSH spec"),
			Index: 3,
			State: bmeventlog.Finished,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 4,
			Task:  fmt.Sprintf("Starting agent services"),
			Index: 4,
			State: bmeventlog.Started,
		}))
		Expect(eventLogger.LoggedEvents).To(ContainElement(bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 4,
			Task:  fmt.Sprintf("Starting agent services"),
			Index: 4,
			State: bmeventlog.Finished,
		}))

		Expect(eventLogger.LoggedEvents).To(HaveLen(6))
	})

	Context("when starting SSH tunnel fails", func() {
		BeforeEach(func() {
			fakeSSHTunnel.SetStartBehavior(errors.New("fake-ssh-tunnel-start-error"), nil)
		})

		It("returns an error", func() {
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
		})
	})

	Context("when starting registry fails", func() {
		BeforeEach(func() {
			fakeRegistryServer.SetStartBehavior(errors.New("fake-registry-start-error"), nil)
		})

		It("returns an error", func() {
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-start-error"))
		})
	})

	Context("when waiting for the agent fails", func() {
		BeforeEach(func() {
			fakeAgentPingRetryStrategy.TryErr = errors.New("fake-ping-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-ping-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 4,
				Task:  fmt.Sprintf("Waiting for the agent"),
				Index: 2,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   4,
				Task:    fmt.Sprintf("Waiting for the agent"),
				Index:   2,
				State:   bmeventlog.Failed,
				Message: "fake-ping-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(2))
		})
	})

	Context("when updating instance fails", func() {
		BeforeEach(func() {
			fakeInstanceUpdater.UpdateErr = errors.New("fake-update-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-update-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 4,
				Task:  fmt.Sprintf("Applying micro BOSH spec"),
				Index: 3,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   4,
				Task:    fmt.Sprintf("Applying micro BOSH spec"),
				Index:   3,
				State:   bmeventlog.Failed,
				Message: "fake-update-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(4))
		})
	})

	Context("when starting agent services fails", func() {
		BeforeEach(func() {
			fakeInstanceUpdater.StartErr = errors.New("fake-start-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-start-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 4,
				Task:  fmt.Sprintf("Starting agent services"),
				Index: 4,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   4,
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
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid", fakeInstanceUpdater)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
		})
	})
})
