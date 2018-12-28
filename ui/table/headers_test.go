package table_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("Headers", func() {

	Describe("KeyifyHeader", func() {
		It("should convert alphanumeric to lowercase ", func() {
			keyifyHeader := table.KeyifyHeader("Header1")
			Expect(keyifyHeader).To(Equal("header1"))
		})

		Context("given a header that only contains non-alphanumeric and alphanumeric", func() {
			It("should non-alphanumeric to underscore", func() {
				keyifyHeader := table.KeyifyHeader("FOO!@AND#$BAR")
				Expect(keyifyHeader).To(Equal("foo_and_bar"))
			})
		})

		Context("given a header that only contains non-alphanumeric", func() {
			It("should convert to underscore", func() {
				keyifyHeader := table.KeyifyHeader("!@#$")
				Expect(keyifyHeader).To(Equal("_"))
			})

			It("should convert empty header to underscore", func() {
				keyifyHeader := table.KeyifyHeader("")
				Expect(keyifyHeader).To(Equal("_"))
			})
		})

	})

	Describe("SetColumnVisibility", func() {
		Context("when given a header that does not exist", func() {
			It("should return an error", func() {
				t := table.Table{
					Header: []table.Header{table.NewHeader("header1")},
				}

				err := t.SetColumnVisibility([]table.Header{table.NewHeader("non-matching-header")})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("SetColumnVisibilityFiltered", func() {
		Context("when given a header that does not exist", func() {
			It("should return an error", func() {
				t := table.Table{
					Header: []table.Header{table.NewHeader("header1")},
				}

				filteredHeader := []table.Header{
					table.NewHeader("non-matching-header-1"),
					table.NewHeader("non-matching-header-2"),
				}

				err := t.SetColumnVisibilityFiltered([]table.Header{table.NewHeader("non-matching-header")}, filteredHeader)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when filterHeader list has a matching table header name but not the column header", func() {
			It("should raise error", func() {
				t := table.Table{
					Header: []table.Header{table.NewHeader("header1")},
				}

				filteredHeader := []table.Header{
					table.NewHeader("header1"),
					table.NewHeader("non-matching-header-2"),
				}

				err := t.SetColumnVisibilityFiltered([]table.Header{table.NewHeader("non-matching-header")}, filteredHeader)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when filterHeader list has a matching column header name", func() {
			It("should not raise error", func() {
				t := table.Table{
					Header: []table.Header{table.NewHeader("header1")},
				}

				filteredHeader := []table.Header{
					table.NewHeader("matching-filter-header"),
					table.NewHeader("non-matching-header-2"),
				}

				err := t.SetColumnVisibilityFiltered([]table.Header{table.NewHeader("matching-filter-header")}, filteredHeader)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when table header has a matching column header name", func() {
			It("should not raise error", func() {
				t := table.Table{
					Header: []table.Header{table.NewHeader("matching-table-header")},
				}

				filteredHeader := []table.Header{
					table.NewHeader("non-matching-header"),
					table.NewHeader("non-matching-header-2"),
				}

				err := t.SetColumnVisibilityFiltered([]table.Header{table.NewHeader("matching-table-header")}, filteredHeader)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
