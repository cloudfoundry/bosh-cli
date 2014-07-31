package cmd_test

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cmd.Factory", func() {
	var factory cmd.Factory
	var logger boshlog.Logger
	var filesystem boshsys.FileSystem

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelDebug)
		factory = cmd.NewFactory(logger)
		filesystem = boshsys.NewOsFileSystem(logger)
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
		Expect(filesystem).ToNot(BeNil())
	})

	Context("passing correct command name", func() {
		It("has deployment command", func() {
			cmd, err := factory.CreateCommand("deployment")
			Expect(err).ToNot(HaveOccurred())
			Expect(cmd).ToNot(BeNil())
			Expect(cmd.FileSystem()).ToNot(BeNil())
		})
	})

	Context("invalid command name", func() {
		It("returns error on invalid command name", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
