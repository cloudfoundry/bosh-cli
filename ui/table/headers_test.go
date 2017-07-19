package table_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("NewHeader", func() {
	It("converts alphanumeric to lowercase", func() {
		Expect(NewHeader("Header1").Key).To(Equal("header1"))
	})

	It("removes '(' and ')'", func() {
		Expect(NewHeader("Header(1)").Key).To(Equal("header1"))
	})

	It("lowercases alphanum and converts non-alphanum to underscores", func() {
		Expect(NewHeader("!@#$").Key).To(Equal(""))
		Expect(NewHeader("FOO!@AND#$BAR").Key).To(Equal("foo_and_bar"))
	})
})
