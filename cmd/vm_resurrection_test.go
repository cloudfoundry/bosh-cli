package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/director/fakes"
)

var _ = Describe("VMResurrectionCmd", func() {
	var (
		director *fakedir.FakeDirector
		command  VMResurrectionCmd
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		command = NewVMResurrectionCmd(director)
	})

	Describe("Run", func() {
		var (
			opts VMResurrectionOpts
		)

		BeforeEach(func() {
			opts = VMResurrectionOpts{}
		})

		act := func() error { return command.Run(opts) }

		It("enables resurrection", func() {
			opts.Args.Enabled = BoolArg(true)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.EnableResurrectionCallCount()).To(Equal(1))
			Expect(director.EnableResurrectionArgsForCall(0)).To(BeTrue())
		})

		It("disables resurrection", func() {
			opts.Args.Enabled = BoolArg(false)

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
