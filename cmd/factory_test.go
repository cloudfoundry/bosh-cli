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

	bmconfig "github.com/cloudfoundry/bosh-init/config"
	bmui "github.com/cloudfoundry/bosh-init/ui"

	fakebmui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory           Factory
		userConfig        bmconfig.UserConfig
		userConfigService bmconfig.UserConfigService
		fs                boshsys.FileSystem
		ui                bmui.UI
		logger            boshlog.Logger
		uuidGenerator     *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		userConfig = bmconfig.UserConfig{DeploymentManifestPath: "/fake-path/manifest.yml"}
		userConfigService = bmconfig.NewFileSystemUserConfigService("/fake-user-config", fs, logger)
		ui = &fakebmui.FakeUI{}
		uuidGenerator = &fakeuuid.FakeGenerator{}

		factory = NewFactory(
			userConfig,
			userConfigService,
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
		Describe("deployment command", func() {
			It("returns deployment command", func() {
				cmd, err := factory.CreateCommand("deployment")
				Expect(err).ToNot(HaveOccurred())
				Expect(cmd.Name()).To(Equal("deployment"))
			})
		})

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
