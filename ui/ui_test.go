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

	Context("#Say", func() {
		It("prints what is said to std out", func() {
			ui.Say("hey")
			Expect(stdOut.String()).To(ContainSubstring("hey"))
		})
	})

	Context("#Error", func() {
		It("prints what is errored to std err", func() {
			ui.Error("fake error")
			Expect(stdErr.String()).To(ContainSubstring("fake error"))
		})
	})
})
