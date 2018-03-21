package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("DiffConfigTable", func() {
	var (
		ui   *fakeui.FakeUI
		opts DiffConfigOpts
		diff Diff
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		opts = DiffConfigOpts{
			FromID: "1",
			ToID:   "2",
		}
		lines := [][]interface{}{
			{"some line that stayed", nil},
			{"some line that was added", "added"},
			{"some line that was removed", "removed"},
		}
		diff = NewDiff(lines)
	})

	Describe("Print", func() {
		It("shows diff config as transposed tabled", func() {
			NewConfigDiffTable(diff, opts, ui).Print()

			Expect(ui.Table).To(Equal(
				boshtbl.Table{
					Content: "",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("From ID"),
						boshtbl.NewHeader("To ID"),
						boshtbl.NewHeader("Diff"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("1"),
							boshtbl.NewValueString("2"),
							boshtbl.NewValueString("  some line that stayed\n+ some line that was added\n- some line that was removed\n"),
						},
					},

					Notes: []string{},

					FillFirstColumn: true,

					Transpose: true,
				}))
		})
	})

})
