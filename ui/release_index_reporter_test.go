package ui_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/ui"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("ReleaseIndexReporter", func() {
	var (
		testUI   *testui.Ui
		reporter ReleaseIndexReporter
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		reporter = NewReleaseIndexReporter(testUI)
	})

	Describe("ReleaseIndexAdded", func() {
		It("prints failed msg", func() {
			reporter.ReleaseIndexAdded("name", "desc", errors.New("err"))
			Expect(testUI.Errors).To(Equal([]string{"Failed adding name release 'desc'"}))
		})

		It("prints finished msg", func() {
			reporter.ReleaseIndexAdded("name", "desc", nil)
			Expect(testUI.Said).To(Equal([]string{"Added name release 'desc'"}))
		})
	})
})
