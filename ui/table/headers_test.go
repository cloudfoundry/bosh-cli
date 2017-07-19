package table_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("KeyifyHeader", func() {
	It("converts alphanumeric to lowercase", func() {
		Expect(KeyifyHeader("Header1")).To(Equal("header1"))
	})

	It("removes '(' and ')'", func() {
		Expect(KeyifyHeader("Header(1)")).To(Equal("header1"))
	})

	It("lowercases alphanum and converts non-alphanum to underscores", func() {
		Expect(KeyifyHeader("!@#$")).To(Equal(""))
		Expect(KeyifyHeader("FOO!@AND#$BAR")).To(Equal("foo_and_bar"))
	})
})

var _ = Describe("SetColumnVisibility", func() {
	It("returns an error when given a header that does not exist", func() {
		t := Table{
			Header: []Header{
				NewHeader("header1"),
				NewHeader("header2"),
			},
		}
		err := t.SetColumnVisibility([]Header{NewHeader("non-matching-header")})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Failed to find header 'non_matching_header' (found headers: 'header1', 'header2')"))
	})
})
