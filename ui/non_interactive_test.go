package ui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	. "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("NonInteractiveUI", func() {
	var (
		parentUI *fakeui.FakeUI
		ui       UI
	)

	BeforeEach(func() {
		parentUI = &fakeui.FakeUI{}
		ui = NewNonInteractiveUI(parentUI)
	})

	Describe("ErrorLinef", func() {
		It("delegates to the parent UI", func() {
			ui.ErrorLinef("fake-error-line")
			Expect(parentUI.Errors).To(Equal([]string{"fake-error-line"}))
		})
	})

	Describe("PrintLinef", func() {
		It("delegates to the parent UI", func() {
			ui.PrintLinef("fake-line")
			Expect(parentUI.Said).To(Equal([]string{"fake-line"}))
		})
	})

	Describe("BeginLinef", func() {
		It("delegates to the parent UI", func() {
			ui.BeginLinef("fake-start")
			Expect(parentUI.Said).To(Equal([]string{"fake-start"}))
		})
	})

	Describe("EndLinef", func() {
		It("delegates to the parent UI", func() {
			ui.EndLinef("fake-end")
			Expect(parentUI.Said).To(Equal([]string{"fake-end"}))
		})
	})

	Describe("PrintBlock", func() {
		It("delegates to the parent UI", func() {
			ui.PrintBlock([]byte("block"))
			Expect(parentUI.Blocks).To(Equal([]string{"block"}))
		})
	})

	Describe("PrintErrorBlock", func() {
		It("delegates to the parent UI", func() {
			ui.PrintErrorBlock("block")
			Expect(parentUI.Blocks).To(Equal([]string{"block"}))
		})
	})

	Describe("PrintTable", func() {
		It("delegates to the parent UI", func() {
			table := Table{
				Content: "things",
				Header:  []Header{NewHeader("header1")},
			}

			ui.PrintTable(table)

			Expect(parentUI.Table).To(Equal(table))
		})
	})

	Describe("PrintTableFiltered", func() {
		It("delegates to the parent UI", func() {
			table := Table{
				Content: "things",
				Header:  []Header{NewHeader("header1")},
			}
			filteredHeader := make([]Header, 0)

			ui.PrintTableFiltered(table, filteredHeader)

			Expect(parentUI.Table).To(Equal(table))
		})
	})

	Describe("AskForText", func() {
		It("panics", func() {
			Expect(func() { _, _ = ui.AskForText("") }).To(Panic()) //nolint:errcheck
		})
	})

	Describe("AskForTextWithDefaultValue", func() {
		It("panics", func() {
			Expect(func() { _, _ = ui.AskForTextWithDefaultValue("", "") }).To(Panic()) //nolint:errcheck
		})
	})

	Describe("AskForPassword", func() {
		It("panics", func() {
			Expect(func() { _, _ = ui.AskForPassword("") }).To(Panic()) //nolint:errcheck
		})
	})

	Describe("AskForChoice", func() {
		It("panics", func() {
			Expect(func() { _, _ = ui.AskForChoice("", nil) }).To(Panic()) //nolint:errcheck
		})
	})

	Describe("AskForConfirmation", func() {
		It("responds affirmatively with no error", func() {
			Expect(ui.AskForConfirmation()).To(BeNil())
		})
	})

	Describe("AskForConfirmationWithLabel", func() {
		It("responds affirmatively with no error", func() {
			Expect(ui.AskForConfirmationWithLabel("")).To(BeNil())
		})
	})

	Describe("IsInteractive", func() {
		It("returns false", func() {
			Expect(ui.IsInteractive()).To(BeFalse())
		})
	})

	Describe("Flush", func() {
		It("delegates to the parent UI", func() {
			ui.Flush()
			Expect(parentUI.Flushed).To(BeTrue())
		})
	})
})
