package cmd_test

import (
	"fmt"

	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	fakes "github.com/cloudfoundry/bosh-micro-cli/cmd/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var runner *cmd.Runner
	var factory *fakes.FakeFactory
	var fakeCommand *fakes.FakeCommand

	BeforeEach(func() {
		fakeCommand = fakes.NewFakeCommand("deployment")
		factory = &fakes.FakeFactory{PresetCommand: fakeCommand}
	})

	Context("#Run", func() {
		Context("valid args", func() {
			BeforeEach(func() {
				runner = cmd.NewRunner(factory)
			})

			It("extracts command name from the arguments", func() {
				err := runner.Run([]string{"deployment", "/fake/manifest_path"})
				Expect(err).To(BeNil())
				Expect(factory.CommandName).To(Equal("deployment"))
			})

			It("creates and run a non nil Command with remaining args", func() {
				err := runner.Run([]string{"deployment", "/fake/manifest_path"})
				Expect(err).To(BeNil())
				Expect(factory.CommandName).To(Equal("deployment"))
				Expect(factory.PresetCommand).ToNot(BeNil())
				Expect(factory.PresetCommand.GetArgs()).To(Equal([]string{"/fake/manifest_path"}))
			})
		})

		Context("invalid args", func() {
			BeforeEach(func() {
				runner = cmd.NewRunner(factory)
			})

			It("fails with error with nil args", func() {
				err := runner.Run(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("Invalid args, cannot be nil"))
				Expect(factory.CommandName).To(Equal(""))
			})

			It("fails with error with empty args", func() {
				err := runner.Run([]string{})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("Invalid args, cannot be empty"))
				Expect(factory.CommandName).To(Equal(""))
			})

			Context("unknown command name", func() {
				var fakeCommandName string

				BeforeEach(func() {
					fakeCommandName = "fake-command-name"
					factory.PresetError = fmt.Errorf("Failed creating command with name: %s", fakeCommandName)
					runner = cmd.NewRunner(factory)
				})

				It("fails with error with unknown command name", func() {
					err := runner.Run([]string{"fake-command-name", "/fake/manifest_path"})
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Failed creating command with name: %s", fakeCommandName)))
					Expect(factory.CommandName).To(Equal("fake-command-name"))
				})
			})
		})
	})
})
