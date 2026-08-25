package ui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/ui"
	. "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("NonTTYUI", func() {
	var (
		testUI *testui.Ui
		ui     UI
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		ui = NewNonTTYUI(testUI)
	})

	Describe("ErrorLinef", func() {
		It("includes in Lines", func() {
			ui.ErrorLinef("fake-line1")
			Expect(testUI.Said).To(BeEmpty())
			Expect(testUI.Errors).To(Equal([]string{"fake-line1"}))
		})
	})

	Describe("PrintLinef", func() {
		It("does not include in Lines", func() {
			ui.PrintLinef("fake-line1")
			Expect(testUI.Said).To(BeEmpty())
			Expect(testUI.Errors).To(BeEmpty())
		})
	})

	Describe("BeginLinef", func() {
		It("does not include in Lines", func() {
			ui.BeginLinef("fake-line1")
			Expect(testUI.Said).To(BeEmpty())
			Expect(testUI.Errors).To(BeEmpty())
		})
	})

	Describe("EndLinef", func() {
		It("does not include in Lines", func() {
			ui.EndLinef("fake-line1")
			Expect(testUI.Said).To(BeEmpty())
			Expect(testUI.Errors).To(BeEmpty())
		})
	})

	Describe("PrintBlock", func() {
		It("delegates to the parent UI", func() {
			ui.PrintBlock([]byte("block"))
			Expect(testUI.Blocks).To(Equal([]string{"block"}))
		})
	})

	Describe("PrintErrorBlock", func() {
		It("delegates to the parent UI", func() {
			ui.PrintBlock([]byte("block"))
			Expect(testUI.Blocks).To(Equal([]string{"block"}))
		})
	})

	Describe("PrintTable", func() {
		It("delegates to the parent UI with re-configured table", func() {
			ui.PrintTable(Table{
				Title:  "title",
				Header: []Header{NewHeader("header1")},

				Notes:   []string{"note1"},
				Content: "things",

				SortBy: []ColumnSort{{Column: 1}},

				Sections: []Section{
					{
						FirstColumn: ValueString{S: "section1"},
						Rows:        [][]Value{{ValueString{S: "row1"}}},
					},
				},

				Rows: [][]Value{{ValueString{S: "row1"}}},

				FillFirstColumn: false,
				BackgroundStr:   "-",
				BorderStr:       "",
			})

			Expect(testUI.Table).To(Equal(Table{
				Title: "",
				Header: []Header{
					{Key: "header1", Title: "header1", Hidden: false},
				},
				HeaderFormatFunc: nil,

				Notes:   nil,
				Content: "",

				SortBy: []ColumnSort{{Column: 1}},

				Sections: []Section{
					{
						FirstColumn: ValueString{S: "section1"},
						Rows:        [][]Value{{ValueString{S: "row1"}}},
					},
				},

				Rows: [][]Value{{ValueString{S: "row1"}}},

				FillFirstColumn: true,
				DataOnly:        true,
				BackgroundStr:   "-",
				BorderStr:       "\t",
			}))
		})
	})

	Describe("IsInteractive", func() {
		It("delegates to the parent UI", func() {
			testUI.Interactive = true
			Expect(ui.IsInteractive()).To(BeTrue())

			testUI.Interactive = false
			Expect(ui.IsInteractive()).To(BeFalse())
		})
	})

	Describe("Flush", func() {
		It("delegates to the parent UI", func() {
			ui.Flush()
			Expect(testUI.Flushed).To(BeTrue())
		})
	})
})
