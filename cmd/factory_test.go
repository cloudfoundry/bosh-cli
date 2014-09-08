package cmd_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakeconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory       Factory
		config        bmconfig.Config
		configService *fakeconfig.FakeService
		filesystem    boshsys.FileSystem
		ui            bmui.UI
		extractor     boshcmd.Compressor
		logger        boshlog.Logger
		uuidGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		config = bmconfig.Config{Deployment: "/fake-path/manifest.yml"}
		configService = &fakeconfig.FakeService{}
		filesystem = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		uuidGenerator = &fakeuuid.FakeGenerator{}

		factory = NewFactory(
			config,
			configService,
			filesystem,
			ui,
			logger,
			uuidGenerator,
		)
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
	})

	Context("passing correct command name", func() {
		Describe("deployment command", func() {
			It("returns a deployment command", func() {
				cmd, err := factory.CreateCommand("deployment")
				Expect(err).ToNot(HaveOccurred())
				Expect(cmd).To(Equal(NewDeploymentCmd(
					ui,
					config,
					configService,
					filesystem,
					uuidGenerator,
					logger,
				)))
			})
		})

		Describe("deploy command", func() {
			It("returns a  deploy command", func() {
				releaseValidator := fakebmrel.NewFakeValidator()
				releaseCompiler := fakebmcomp.NewFakeReleaseCompiler()

				cmd, err := factory.CreateCommand("deploy")
				Expect(err).ToNot(HaveOccurred())
				Expect(cmd).To(BeAssignableToTypeOf(NewDeployCmd(
					ui,
					config,
					filesystem,
					extractor,
					releaseValidator,
					releaseCompiler,
					logger,
				)))
			})
		})
	})

	Context("invalid command name", func() {
		It("returns error on invalid command name", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
