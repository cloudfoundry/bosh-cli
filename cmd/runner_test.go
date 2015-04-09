package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	bmcmd "github.com/cloudfoundry/bosh-init/cmd"

	fakebmcmd "github.com/cloudfoundry/bosh-init/cmd/fakes"
	fakebmui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("Runner", func() {
	var (
		runner      *bmcmd.Runner
		factory     *fakebmcmd.FakeFactory
		fakeCommand *fakebmcmd.FakeCommand
		fakeStage   *fakebmui.FakeStage
	)

	BeforeEach(func() {
		fakeCommand = fakebmcmd.NewFakeCommand("deployment")
		factory = &fakebmcmd.FakeFactory{PresetCommand: fakeCommand}
		fakeStage = fakebmui.NewFakeStage()
	})

	JustBeforeEach(func() {
		runner = bmcmd.NewRunner(factory)
	})

	Context("Run", func() {
		Context("valid args", func() {
			It("extracts command name from the arguments", func() {
				err := runner.Run(fakeStage, "deployment", "/fake/manifest_path")
				Expect(err).ToNot(HaveOccurred())
				Expect(factory.CommandName).To(Equal("deployment"))
			})

			It("creates and run a non nil Command with remaining args", func() {
				err := runner.Run(fakeStage, "deployment", "/fake/manifest_path")
				Expect(err).ToNot(HaveOccurred())
				Expect(factory.CommandName).To(Equal("deployment"))
				Expect(factory.PresetCommand).ToNot(BeNil())
				Expect(factory.PresetCommand.GetArgs()).To(Equal([]string{"/fake/manifest_path"}))
			})
		})

		Context("invalid args", func() {
			It("fails with error with empty args", func() {
				err := runner.Run(fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid usage: No command specified"))
				Expect(factory.CommandName).To(Equal(""))
			})

			Context("unknown command name", func() {
				var fakeCommandName string

				BeforeEach(func() {
					fakeCommandName = "fake-command-name"
					factory.PresetError = fmt.Errorf("Failed creating command with name: %s", fakeCommandName)
				})

				It("fails with error with unknown command name", func() {
					err := runner.Run(fakeStage, "fake-command-name", "/fake/manifest_path")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Command '%s' unknown", fakeCommandName)))
					Expect(factory.CommandName).To(Equal("fake-command-name"))
				})
			})
		})
	})
})
