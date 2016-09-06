package ui_test

import (
	"encoding/json"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("JSONUI", func() {
	var (
		parentUI *fakeui.FakeUI
		ui       UI
	)

	BeforeEach(func() {
		parentUI = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		ui = NewJSONUI(parentUI, logger)
	})

	type tableResp struct {
		Content string
		Header  []string
		Rows    [][]string
		Notes   []string
	}

	type uiResp struct {
		Tables []tableResp
		Blocks []string
		Lines  []string
	}

	finalOutput := func() uiResp {
		ui.Flush()

		var val uiResp

		err := json.Unmarshal([]byte(parentUI.Blocks[0]), &val)
		if err != nil {
			panic("Unmarshaling")
		}

		return val
	}

	Describe("ErrorLinef", func() {
		It("includes in Lines", func() {
			ui.ErrorLinef("fake-line1")
			ui.ErrorLinef("fake-line2")
			Expect(finalOutput()).To(Equal(uiResp{
				Lines: []string{"fake-line1", "fake-line2"},
			}))
		})
	})

	Describe("PrintLinef", func() {
		It("includes in Lines", func() {
			ui.PrintLinef("fake-line1")
			ui.PrintLinef("fake-line2")
			Expect(finalOutput()).To(Equal(uiResp{
				Lines: []string{"fake-line1", "fake-line2"},
			}))
		})
	})

	Describe("BeginLinef", func() {
		It("includes in Lines", func() {
			ui.BeginLinef("fake-line1")
			ui.BeginLinef("fake-line2")
			Expect(finalOutput()).To(Equal(uiResp{
				Lines: []string{"fake-line1", "fake-line2"},
			}))
		})
	})

	Describe("EndLinef", func() {
		It("includes in Lines", func() {
			ui.EndLinef("fake-line1")
			ui.EndLinef("fake-line2")
			Expect(finalOutput()).To(Equal(uiResp{
				Lines: []string{"fake-line1", "fake-line2"},
			}))
		})
	})

	Describe("PrintBlock", func() {
		It("includes in Blocks", func() {
			ui.PrintBlock("fake-block1")
			ui.PrintBlock("fake-block2")
			Expect(finalOutput()).To(Equal(uiResp{
				Blocks: []string{"fake-block1", "fake-block2"},
			}))
		})
	})

	Describe("PrintErrorBlock", func() {
		It("includes in Blocks", func() {
			ui.PrintErrorBlock("fake-block1")
			ui.PrintErrorBlock("fake-block2")
			Expect(finalOutput()).To(Equal(uiResp{
				Blocks: []string{"fake-block1", "fake-block2"},
			}))
		})
	})

	Describe("PrintTable", func() {
		It("includes in Tables", func() {
			table := Table{
				Content: "things",
				Header:  []string{"Header1", "Header2"},

				Rows: [][]Value{
					{ValueString{"r1c1"}, ValueString{"r1c2"}},
					{ValueString{"r2c1"}, ValueString{"r2c2"}},
				},

				Notes: []string{"note1", "note2"},
			}

			table2 := Table{
				Content: "things2",
			}

			ui.PrintTable(table)
			ui.PrintTable(table2)

			Expect(finalOutput()).To(Equal(uiResp{
				Tables: []tableResp{
					{
						Content: "things",
						Header:  []string{"Header1", "Header2"},
						Rows:    [][]string{{"r1c1", "r1c2"}, {"r2c1", "r2c2"}},
						Notes:   []string{"note1", "note2"},
					},
					{
						Content: "things2",
					},
				},
			}))
		})

		It("includes HeaderVals in Tables", func() {
			table := Table{
				Content: "things",
				HeaderVals: []Value{
					ValueString{"Header1"},
					ValueString{"Header2"},
				},

				Rows: [][]Value{
					{ValueString{"r1c1"}, ValueString{"r1c2"}},
					{ValueString{"r2c1"}, ValueString{"r2c2"}},
				},

				Notes: []string{"note1", "note2"},
			}

			table2 := Table{
				Content: "things2",
			}

			ui.PrintTable(table)
			ui.PrintTable(table2)

			Expect(finalOutput()).To(Equal(uiResp{
				Tables: []tableResp{
					{
						Content: "things",
						Header:  []string{"Header1", "Header2"},
						Rows:    [][]string{{"r1c1", "r1c2"}, {"r2c1", "r2c2"}},
						Notes:   []string{"note1", "note2"},
					},
					{
						Content: "things2",
					},
				},
			}))
		})

		It("includes in Tables when table has sections and fills in first column", func() {
			table := Table{
				Content: "things",
				Header:  []string{"Header1", "Header2"},

				Sections: []Section{
					{
						FirstColumn: ValueString{"first-col"},
						Rows: [][]Value{
							{ValueString{""}, ValueString{"r1c2"}},
							{ValueString{""}, ValueString{"r2c2"}},
						},
					},
				},

				Notes: []string{"note1", "note2"},
			}

			ui.PrintTable(table)

			Expect(finalOutput()).To(Equal(uiResp{
				Tables: []tableResp{
					{
						Content: "things",
						Header:  []string{"Header1", "Header2"},
						Rows: [][]string{
							{"first-col", "r1c2"},
							{"first-col", "r2c2"},
						},
						Notes: []string{"note1", "note2"},
					},
				},
			}))
		})
	})

	Describe("AskForText", func() {
		It("panics", func() {
			Expect(func() { ui.AskForText("") }).To(Panic())
		})
	})

	Describe("AskForPassword", func() {
		It("panics", func() {
			Expect(func() { ui.AskForPassword("") }).To(Panic())
		})
	})

	Describe("AskForChoice", func() {
		It("panics", func() {
			Expect(func() { ui.AskForChoice("", nil) }).To(Panic())
		})
	})

	Describe("AskForConfirmation", func() {
		It("panics", func() {
			Expect(func() { ui.AskForConfirmation() }).To(Panic())
		})
	})

	Describe("IsInteractive", func() {
		It("delegates to the parent UI", func() {
			parentUI.Interactive = true
			Expect(ui.IsInteractive()).To(BeTrue())

			parentUI.Interactive = false
			Expect(ui.IsInteractive()).To(BeFalse())
		})
	})

	Describe("Flush", func() {
		It("does not output anything when nothing was recorded", func() {
			ui.Flush()
			Expect(parentUI.Said).To(BeEmpty())
		})

		It("outputs everything when something was recorded", func() {
			ui.PrintLinef("fake-line1")
			ui.Flush()
			Expect(parentUI.Blocks[0]).To(Equal(`{
    "Tables": null,
    "Blocks": null,
    "Lines": [
        "fake-line1"
    ]
}`))
		})
	})
})
