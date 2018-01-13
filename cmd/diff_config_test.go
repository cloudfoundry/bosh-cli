package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("DiffConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  DiffConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewDiffConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts DiffConfigOpts
		)

		BeforeEach(func() {
			opts = DiffConfigOpts{
				Args: DiffConfigArgs{
					FromID: "1",
					ToID:   "2",
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("diff two configs", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.DiffConfigByIDCallCount()).To(Equal(1))

			from, to := director.DiffConfigByIDArgsForCall(0)
			Expect(from).To(Equal("1"))
			Expect(to).To(Equal("2"))
		})

		It("returns an error if diffing failed", func() {
			director.DiffConfigByIDReturns(boshdir.ConfigDiff{}, errors.New("Fetching diff result"))

			err := act()
			Expect(err).To(HaveOccurred())
		})

		It("gets the diff from the two configs", func() {
			diff := [][]interface{}{
				[]interface{}{"some line that stayed", nil},
				[]interface{}{"some line that was added", "added"},
				[]interface{}{"some line that was removed", "removed"},
			}

			expectedDiff := boshdir.NewConfigDiff(diff)
			director.DiffConfigByIDReturns(expectedDiff, nil)
			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(director.DiffConfigByIDCallCount()).To(Equal(1))
			Expect(ui.Said).To(ContainElement("  some line that stayed\n"))
			Expect(ui.Said).To(ContainElement("+ some line that was added\n"))
			Expect(ui.Said).To(ContainElement("- some line that was removed\n"))
		})
	})
})
