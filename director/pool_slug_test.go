package director_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/director"
)

var _ = Describe("PoolSlug", func() {
	Describe("Name", func() {
		It("returns name", func() {
			Expect(NewPoolSlug("name").String()).To(Equal("name"))
		})
	})

	Describe("String", func() {
		It("returns name", func() {
			Expect(NewPoolSlug("name").String()).To(Equal("name"))
		})
	})

	Describe("UnmarshalFlag", func() {
		var (
			slug *PoolSlug
		)

		BeforeEach(func() {
			slug = &PoolSlug{}
		})

		It("populates slug", func() {
			err := slug.UnmarshalFlag("name")
			Expect(err).ToNot(HaveOccurred())
			Expect(*slug).To(Equal(NewPoolSlug("name")))
		})

		It("returns an error if name is empty", func() {
			err := slug.UnmarshalFlag("")
			Expect(err).To(Equal(errors.New("Expected non-empty pool name")))
		})
	})
})
