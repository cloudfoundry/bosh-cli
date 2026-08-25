package ui_test

import (
	"bytes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/ui"
	. "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("IndentingUI", func() {
	var (
		uiOut, uiErr *bytes.Buffer
		testUI       *testui.Ui
		parentUI     UI
		ui           UI
	)

	BeforeEach(func() {
		uiOut = bytes.NewBufferString("")
		uiErr = bytes.NewBufferString("")

		logger := boshlog.NewLogger(boshlog.LevelNone)
		parentUI = NewWriterUI(uiOut, uiErr, logger)
		testUI = &testui.Ui{}
	})

	JustBeforeEach(func() {
		ui = NewIndentingUI(parentUI)
	})

	Describe("ErrorLinef", func() {
		It("delegates to the parent UI with an indent", func() {
			ui.ErrorLinef("fake-error-line")
			Expect(uiErr.String()).To(ContainSubstring("  fake-error-line\n"))
			Expect(uiOut.String()).To(BeEmpty())
		})
	})

	Describe("PrintLinef", func() {
		It("delegates to the parent UI with an indent", func() {
			ui.PrintLinef("fake-line")
			Expect(uiOut.String()).To(ContainSubstring("  fake-line\n"))
			Expect(uiErr.String()).To(BeEmpty())
		})
	})

	Describe("BeginLinef", func() {
		It("delegates to the parent UI with an indent", func() {
			ui.BeginLinef("fake-start")
			Expect(uiOut.String()).To(ContainSubstring("  fake-start"))
			Expect(uiErr.String()).To(BeEmpty())
		})
	})

	Describe("EndLinef", func() {
		It("delegates to the parent UI", func() {
			ui.EndLinef("fake-end")
			Expect(uiOut.String()).To(ContainSubstring("fake-end\n"))
			Expect(uiErr.String()).To(BeEmpty())
		})
	})

	Describe("PrintBlock", func() {
		BeforeEach(func() {
			parentUI = testUI
		})

		It("delegates to the parent UI", func() {
			ui.PrintBlock([]byte("block"))
			Expect(testUI.Blocks).To(Equal([]string{"block"}))
		})
	})

	Describe("PrintErrorBlock", func() {
		BeforeEach(func() {
			parentUI = testUI
		})

		It("delegates to the parent UI", func() {
			ui.PrintBlock([]byte("block"))
			Expect(testUI.Blocks).To(Equal([]string{"block"}))
		})
	})

	Describe("PrintTable", func() {
		BeforeEach(func() {
			parentUI = testUI
		})

		It("delegates to the parent UI", func() {
			table := Table{
				Content: "things",
				Header:  []Header{NewHeader("header1")},
			}

			ui.PrintTable(table)

			Expect(testUI.Table).To(Equal(table))
		})
	})

	Describe("PrintTableFiltered", func() {
		BeforeEach(func() {
			parentUI = testUI
		})

		It("delegates to the parent UI", func() {
			table := Table{
				Content: "things",
				Header:  []Header{NewHeader("header1")},
			}
			filteredHeader := []Header{}

			ui.PrintTableFiltered(table, filteredHeader)

			Expect(testUI.Table).To(Equal(table))
		})
	})

	Describe("IsInteractive", func() {
		BeforeEach(func() {
			parentUI = testUI
		})

		It("delegates to the parent UI", func() {
			testUI.Interactive = true
			Expect(ui.IsInteractive()).To(BeTrue())

			testUI.Interactive = false
			Expect(ui.IsInteractive()).To(BeFalse())
		})
	})

	Describe("Flush", func() {
		BeforeEach(func() {
			parentUI = testUI
		})

		It("delegates to the parent UI", func() {
			ui.Flush()
			Expect(testUI.Flushed).To(BeTrue())
		})
	})
})
