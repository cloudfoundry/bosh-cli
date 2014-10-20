package microdeployer_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/sshtunnel"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakeregistry "github.com/cloudfoundry/bosh-micro-cli/registry/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/sshtunnel/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/vm/fakes"
)

var _ = Describe("MicroDeployer", func() {
	var (
		microDeployer        Deployer
		fakeVMManagerFactory *fakebmvm.FakeManagerFactory
		fakeVMManager        *fakebmvm.FakeManager
		cloud                *fakebmcloud.FakeCloud
		deployment           bmdepl.Deployment
		registry             bmdepl.Registry
		fakeRegistryServer   *fakeregistry.FakeServer
		fakeSSHTunnel        *fakebmsshtunnel.FakeTunnel
		fakeSSHTunnelFactory *fakebmsshtunnel.FakeFactory
		sshTunnelConfig      bmdepl.SSHTunnel
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
		microDeployer = NewMicroDeployer(fakeVMManagerFactory, fakeSSHTunnelFactory, fakeRegistryServer, logger)
	})

	It("starts the registry", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, "fake-stemcell-cid")
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
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVMManager.CreateVMInput).To(Equal(
			fakebmvm.CreateVMInput{
				StemcellCID: "fake-stemcell-cid",
				Deployment:  deployment,
			},
		))
	})

	It("starts the SSH tunnel", func() {
		err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, "fake-stemcell-cid")
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

	Context("when creating VM fails", func() {
		It("returns an error", func() {
			createVMError := errors.New("fake-create-vm-error")
			fakeVMManager.SetCreateVMBehavior("fake-stemcell-cid", createVMError)
			err := microDeployer.Deploy(cloud, deployment, registry, sshTunnelConfig, "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
		})
	})
})
