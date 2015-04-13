package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	biconfig "github.com/cloudfoundry/bosh-init/config"
	biui "github.com/cloudfoundry/bosh-init/ui"

	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory       Factory
		userConfig    biconfig.UserConfig
		fs            boshsys.FileSystem
		ui            biui.UI
		logger        boshlog.Logger
		uuidGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		userConfig = biconfig.UserConfig{DeploymentManifestPath: "/fake-path/manifest.yml"}
		ui = &fakebiui.FakeUI{}
		uuidGenerator = &fakeuuid.FakeGenerator{}

		factory = NewFactory(
			userConfig,
			fs,
			ui,
			boshtime.NewConcreteService(),
			logger,
			uuidGenerator,
			"/fake-path",
		)
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
	})

	Context("known command name", func() {
		Describe("deploy command", func() {
			It("returns deploy command", func() {
				cmd, err := factory.CreateCommand("deploy")
				Expect(err).ToNot(HaveOccurred())
				Expect(cmd.Name()).To(Equal("deploy"))
			})
		})

		Describe("delete command", func() {
			It("returns delete command", func() {
				cmd, err := factory.CreateCommand("delete")
				Expect(err).ToNot(HaveOccurred())
				Expect(cmd.Name()).To(Equal("delete"))
			})
		})
	})

	Context("unknown command name", func() {
		It("returns error", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
