package cmd_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory          Factory
		config           bmconfig.Config
		configService    *fakeconfig.FakeService
		filesystem       boshsys.FileSystem
		ui               bmui.UI
		extractor        bmtar.Extractor
		releaseValidator *fakebmrel.FakeValidator
	)

	BeforeEach(func() {
		config = bmconfig.Config{}
		configService = &fakeconfig.FakeService{}
		filesystem = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}
		runner := fakesys.NewFakeCmdRunner()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		extractor = bmtar.NewCmdExtractor(runner, logger)
		releaseValidator = fakebmrel.NewFakeValidator()

		factory = NewFactory(config, configService, filesystem, ui, extractor, releaseValidator)
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
	})

	Context("passing correct command name", func() {
		It("has deployment command", func() {
			cmd, err := factory.CreateCommand("deployment")
			Expect(err).ToNot(HaveOccurred())
			Expect(cmd).To(Equal(NewDeploymentCmd(ui, config, configService, filesystem)))
		})

		It("has deploy command", func() {
			cmd, err := factory.CreateCommand("deploy")
			Expect(err).ToNot(HaveOccurred())
			Expect(cmd).To(Equal(NewDeployCmd(ui, config, filesystem, extractor, releaseValidator)))
		})
	})

	Context("invalid command name", func() {
		It("returns error on invalid command name", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
