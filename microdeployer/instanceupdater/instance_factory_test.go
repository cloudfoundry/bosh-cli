package instanceupdater_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instanceupdater"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/microdeployer/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec/fakes"
	fakebminsup "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instanceupdater/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceFactory", func() {
	var (
		instanceFactory            InstanceFactory
		fakeAgentClientFactory     *fakebmagentclient.FakeAgentClientFactory
		fakeTemplatesSpecGenerator *fakebminsup.FakeTemplatesSpecGenerator
		fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
		fs                         *fakesys.FakeFileSystem
		logger                     boshlog.Logger
	)
	BeforeEach(func() {
		fakeAgentClientFactory = fakebmagentclient.NewFakeAgentClientFactory()
		fakeTemplatesSpecGenerator = fakebminsup.NewFakeTemplatesSpecGenerator()
		fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		instanceFactory = NewInstanceFactory(fakeAgentClientFactory, fakeTemplatesSpecGenerator, fakeApplySpecFactory, fs, logger)
	})

	Describe("Create", func() {
		It("creates an instance", func() {
			fakeAgentClient := fakebmagentclient.NewFakeAgentClient()
			fakeAgentClientFactory.CreateAgentClient = fakeAgentClient

			expectedInstance := NewInstance(
				fakeAgentClient,
				fakeTemplatesSpecGenerator,
				fakeApplySpecFactory,
				"fake-mbus-url",
				fs,
				logger,
			)

			instance := instanceFactory.Create("fake-mbus-url")
			Expect(instance).To(Equal(expectedInstance))
		})
	})
})
