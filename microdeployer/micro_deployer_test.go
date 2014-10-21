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
		fakeSSHTunnel.SetStartBehavior(struct{}{}, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel
		logger := boshlog.NewLogger(boshlog.LevelNone)
		eventLogger = fakebmlog.NewFakeEventLogger()
		microDeployer = NewMicroDeployer(fakeVMManagerFactory, fakeSSHTunnelFactory, fakeRegistryServer, eventLogger, logger)
		fakeAgentPingRetryStrategy = fakebmretry.NewFakeRetryStrategy()
	})

	It("starts the registry", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
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
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVMManager.CreateVMInput).To(Equal(
			fakebmvm.CreateVMInput{
				StemcellCID: "fake-stemcell-cid",
				Deployment:  deployment,
			},
		))
	})

	It("starts the SSH tunnel", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
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
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeAgentPingRetryStrategy.TryCalled).To(BeTrue())
	})

	It("logs start and stop events to the eventLogger", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())

		expectedStartEvent := bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 1,
			Task:  fmt.Sprintf("Waiting for the agent"),
			Index: 1,
			State: bmeventlog.Started,
		}

		expectedFinishEvent := bmeventlog.Event{
			Stage: "Deploy Micro BOSH",
			Total: 1,
			Task:  fmt.Sprintf("Waiting for the agent"),
			Index: 1,
			State: bmeventlog.Finished,
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFinishEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(2))
	})

	Context("when waiting for the agent fails", func() {
		BeforeEach(func() {
			fakeAgentPingRetryStrategy.TryErr = errors.New("fake-ping-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-ping-error"))

			expectedStartEvent := bmeventlog.Event{
				Stage: "Deploy Micro BOSH",
				Total: 1,
				Task:  fmt.Sprintf("Waiting for the agent"),
				Index: 1,
				State: bmeventlog.Started,
			}

			expectedFailedEvent := bmeventlog.Event{
				Stage:   "Deploy Micro BOSH",
				Total:   1,
				Task:    fmt.Sprintf("Waiting for the agent"),
				Index:   1,
				State:   bmeventlog.Failed,
				Message: "fake-ping-error",
			}
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(2))
		})
	})

	Context("when creating VM fails", func() {
		It("returns an error", func() {
			createVMError := errors.New("fake-create-vm-error")
			fakeVMManager.SetCreateVMBehavior("fake-stemcell-cid", createVMError)
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, fakeAgentPingRetryStrategy, "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
		})
	})
})
