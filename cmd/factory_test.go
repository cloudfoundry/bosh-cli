package cmd_test

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("cmd.Factory", func() {
	var (
		factory    Factory
		filesystem boshsys.FileSystem
		ui         bmui.UI
	)

	BeforeEach(func() {
		filesystem = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}

		factory = NewFactory(filesystem, ui, "/somepath")
	})

	It("creates a new factory", func() {
		Expect(factory).ToNot(BeNil())
	})

	Context("passing correct command name", func() {
		It("has deployment command", func() {
			cmd, err := factory.CreateCommand("deployment")
			Expect(err).ToNot(HaveOccurred())
			Expect(cmd).To(Equal(NewDeploymentCmd(ui, "/somepath", filesystem)))
		})
	})

	Context("invalid command name", func() {
		It("returns error on invalid command name", func() {
			_, err := factory.CreateCommand("bogus-cmd-name")
			Expect(err).To(HaveOccurred())
		})
	})
})
