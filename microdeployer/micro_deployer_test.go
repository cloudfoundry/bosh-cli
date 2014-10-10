package microdeployer_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakeregistry "github.com/cloudfoundry/bosh-micro-cli/registry/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/vm/fakes"
)

var _ = Describe("MicroDeployer", func() {
	var (
		microDeployer        Deployer
		fakeVMManagerFactory *fakebmvm.FakeManagerFactory
		fakeVMManager        *fakebmvm.FakeManager
		cloud                *fakebmcloud.FakeCloud
		deployment           bmdepl.Deployment
		fakeRegistryServer   *fakeregistry.FakeServer
	)

	BeforeEach(func() {
		deployment = bmdepl.Deployment{
			Registry: bmdepl.Registry{
				Username: "fake-username",
				Password: "fake-password",
				Host:     "fake-host",
				Port:     123,
			},
		}
		cloud = fakebmcloud.NewFakeCloud()
		fakeRegistryServer = fakeregistry.NewFakeServer()
		fakeVMManagerFactory = fakebmvm.NewFakeManagerFactory()
		fakeVMManager = fakebmvm.NewFakeManager()
		fakeVMManagerFactory.SetNewManagerBehavior(cloud, fakeVMManager)
		logger := boshlog.NewLogger(boshlog.LevelNone)
		microDeployer = NewMicroDeployer(fakeVMManagerFactory, fakeRegistryServer, logger)
	})

	It("starts the registry", func() {
		err := microDeployer.Deploy(cloud, deployment, "fake-stemcell-cid")
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
		err := microDeployer.Deploy(cloud, deployment, "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVMManager.CreateVMInput).To(Equal(
			fakebmvm.CreateVMInput{
				StemcellCID: "fake-stemcell-cid",
				Deployment:  deployment,
			},
		))
	})

	Context("when creating VM fails", func() {
		It("returns an error", func() {
			createVMError := errors.New("fake-create-vm-error")
			fakeVMManager.SetCreateVMBehavior("fake-stemcell-cid", createVMError)
			err := microDeployer.Deploy(cloud, deployment, "fake-stemcell-cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
		})
	})
})
