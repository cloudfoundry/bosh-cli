package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("UpdateResurrectionCmd", func() {
	var (
		director *fakedir.FakeDirector
		command  cmd.UpdateResurrectionCmd
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		command = cmd.NewUpdateResurrectionCmd(director)
	})

	Describe("Run", func() {
		var (
			updateResurrectionOpts opts.UpdateResurrectionOpts
		)

		BeforeEach(func() {
			updateResurrectionOpts = opts.UpdateResurrectionOpts{}
		})

		act := func() error { return command.Run(updateResurrectionOpts) }

		It("enables resurrection", func() {
			updateResurrectionOpts.Args.Enabled = opts.BoolArg(true)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.EnableResurrectionCallCount()).To(Equal(1))
			Expect(director.EnableResurrectionArgsForCall(0)).To(BeTrue())
		})

		It("disables resurrection", func() {
			updateResurrectionOpts.Args.Enabled = opts.BoolArg(false)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.EnableResurrectionCallCount()).To(Equal(1))
			Expect(director.EnableResurrectionArgsForCall(0)).To(BeFalse())
		})

		It("returns error if changing resurrection fails", func() {
			director.EnableResurrectionReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
