package cmd_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
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
	fakews "github.com/cloudfoundry/bosh-micro-cli/workspace/fakes"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory       Factory
		config        bmconfig.Config
		configService *fakeconfig.FakeService
		filesystem    boshsys.FileSystem
		ui            bmui.UI
		extractor     bmtar.Extractor
		logger        boshlog.Logger
		workspace     *fakews.FakeWorkspace
	)

	BeforeEach(func() {
		config = bmconfig.Config{}
		configService = &fakeconfig.FakeService{}
		filesystem = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		workspace = fakews.NewFakeWorkspace()
		uuidGenerator := &fakeuuid.FakeGenerator{}

		factory = NewFactory(
			config,
			configService,
			filesystem,
			ui,
			logger,
			workspace,
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
					workspace,
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
				Expect(workspace.LoadCalled).To(BeTrue())
			})

			It("loads a workspace", func() {
				_, err := factory.CreateCommand("deploy")
				Expect(err).ToNot(HaveOccurred())

				Expect(workspace.LoadCalled).To(BeTrue())
			})

			It("errors when a workspace cannot be loaded", func() {
				workspace.LoadError = errors.New("fake-load-workspace-error")
				_, err := factory.CreateCommand("deploy")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Loading workspace"))
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
