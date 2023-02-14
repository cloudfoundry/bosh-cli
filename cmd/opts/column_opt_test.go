package opts_test

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ColumnOpt", func() {

	It("should keyify column", func() {
		var columnOpt ColumnOpt
		err := columnOpt.UnmarshalFlag("Header1")
		Expect(err).ToNot(HaveOccurred())

		Expect(columnOpt.Key).To(Equal("header1"))
		Expect(columnOpt.Hidden).To(BeFalse())
	})
})
