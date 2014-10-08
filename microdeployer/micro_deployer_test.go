package microdeployer_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/vm/fakes"
)

var _ = Describe("MicroDeployer", func() {
	var (
		microDeployer        Deployer
		fakeVMManagerFactory *fakebmvm.FakeManagerFactory
		fakeVMManager        *fakebmvm.FakeManager
		cloud                *fakebmcloud.FakeCloud
		deployment           bmdepl.Deployment
	)

	BeforeEach(func() {
		deployment = bmdepl.Deployment{}
		cloud = fakebmcloud.NewFakeCloud()
		fakeVMManagerFactory = fakebmvm.NewFakeManagerFactory()
		fakeVMManager = fakebmvm.NewFakeManager()
		fakeVMManagerFactory.SetNewManagerBehavior(cloud, fakeVMManager)
		microDeployer = NewMicroDeployer(fakeVMManagerFactory)
	})

	It("creates a VM", func() {
		err := microDeployer.Deploy(cloud, deployment, "fake-stemcell-cid")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVMManager.CreateVMInput).To(Equal(
			fakebmvm.CreateVMInput{
				StemcellCID: "fake-stemcell-cid",
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
