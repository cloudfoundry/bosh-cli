package cmd_test

import (
	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cmd.Factory", func() {
	var factory cmd.Factory
	BeforeEach(func() {
		factory = cmd.NewFactory()
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
	})

	Context("passing correct command name", func() {
		It("has deployment command", func() {
			cmd, err := factory.CreateCommand("deployment")
			Expect(err).ToNot(HaveOccurred())
			Expect(cmd).ToNot(BeNil())
		})
	})

	Context("invalid command name", func() {
		It("returns error on invalid command name", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
