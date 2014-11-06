package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec/fakes"
	fakebmins "github.com/cloudfoundry/bosh-micro-cli/deployer/instance/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceFactory", func() {
	var (
		instanceFactory            Factory
		fakeAgentClientFactory     *fakebmagentclient.FakeAgentClientFactory
		fakeTemplatesSpecGenerator *fakebmins.FakeTemplatesSpecGenerator
		fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
		fs                         *fakesys.FakeFileSystem
		logger                     boshlog.Logger
	)
	BeforeEach(func() {
		fakeAgentClientFactory = fakebmagentclient.NewFakeAgentClientFactory()
		fakeTemplatesSpecGenerator = fakebmins.NewFakeTemplatesSpecGenerator()
		fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		instanceFactory = NewInstanceFactory(fakeAgentClientFactory, fakeTemplatesSpecGenerator, fakeApplySpecFactory, fs, logger)
	})

	Describe("Create", func() {
		It("creates an instance", func() {
			fakeAgentClient := fakebmagentclient.NewFakeAgentClient()
			fakeAgentClientFactory.CreateAgentClient = fakeAgentClient
			fakeCloud := fakebmcloud.NewFakeCloud()

			expectedInstance := NewInstance(
				"fake-vm-cid",
				fakeAgentClient,
				fakeCloud,
				fakeTemplatesSpecGenerator,
				fakeApplySpecFactory,
				"fake-mbus-url",
				fs,
				logger,
			)

			instance := instanceFactory.Create("fake-vm-cid", "fake-mbus-url", fakeCloud)
			Expect(instance).To(Equal(expectedInstance))
		})
	})
})
