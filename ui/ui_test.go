package ui_test

import (
	"bytes"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UI", func() {
	var ui bmui.UI
	var stdOut, stdErr *bytes.Buffer

	BeforeEach(func() {
		stdOut = bytes.NewBufferString("")
		stdErr = bytes.NewBufferString("")

		ui = bmui.NewDefaultUI(stdOut, stdErr)
	})

	Context("#Sayln", func() {
		It("prints what is said to std out with a trailing newline", func() {
			ui.Sayln("hey")
			Expect(stdOut.String()).To(ContainSubstring("hey\n"))
		})
	})

	Context("#Error", func() {
		It("prints what is errored to std err with a trailing newline", func() {
			ui.Error("fake error")
			Expect(stdErr.String()).To(ContainSubstring("fake error\n"))
		})
	})

	Context("#Say", func() {
		It("prints what is said to std out without a trailing newline", func() {
			ui.Say("hey")
			Expect(stdOut.String()).To(ContainSubstring("hey"))
			Expect(stdOut.String()).NotTo(ContainSubstring("\n"))
		})
	})
})
