package director_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/director"
)

var _ = Describe("NewPoolOrInstanceSlug", func() {
	It("populates slug when name is just given", func() {
		slug := NewPoolOrInstanceSlug("name", "")
		Expect(slug.Name()).To(Equal("name"))
		Expect(slug.IndexOrID()).To(Equal(""))
	})

	It("populates slug when name and index-or-id is given", func() {
		slug := NewPoolOrInstanceSlug("name", "id")
		Expect(slug.Name()).To(Equal("name"))
		Expect(slug.IndexOrID()).To(Equal("id"))
	})

	It("panics if name is empty", func() {
		Expect(func() { NewPoolOrInstanceSlug("", "") }).To(Panic())
	})
})

var _ = Describe("NewPoolOrInstanceSlugFromString", func() {
	It("populates slug when name is just given", func() {
		slug, err := NewPoolOrInstanceSlugFromString("name")
		Expect(err).ToNot(HaveOccurred())
		Expect(slug).To(Equal(NewPoolOrInstanceSlug("name", "")))
	})

	It("populates slug when name and index-or-id is given", func() {
		slug, err := NewPoolOrInstanceSlugFromString("name/id")
		Expect(err).ToNot(HaveOccurred())
		Expect(slug).To(Equal(NewPoolOrInstanceSlug("name", "id")))
	})

	It("returns an error if string doesnt have 1 or 2 pieces", func() {
		_, err := NewPoolOrInstanceSlugFromString("")
		Expect(err).To(Equal(errors.New("Expected pool or instance '' to specify non-empty name")))

		_, err = NewPoolOrInstanceSlugFromString("1/2/3")
		Expect(err).To(Equal(errors.New("Expected pool or instance '1/2/3' to be in format 'name' or 'name/id-or-index'")))
	})

	It("returns an error if name is empty", func() {
		_, err := NewPoolOrInstanceSlugFromString("/")
		Expect(err).To(Equal(errors.New("Expected pool or instance '/' to specify non-empty name")))
	})

	It("returns an error if index-or-id is empty", func() {
		_, err := NewPoolOrInstanceSlugFromString("name/")
		Expect(err).To(Equal(errors.New("Expected instance 'name/' to specify non-empty ID or index")))
	})
})

var _ = Describe("PoolOrInstanceSlug", func() {
	Describe("String", func() {
		It("returns name string if id is not set", func() {
			Expect(NewPoolOrInstanceSlug("name", "").String()).To(Equal("name"))
		})

		It("returns name/id string if id is set", func() {
			Expect(NewPoolOrInstanceSlug("name", "id").String()).To(Equal("name/id"))
		})
	})
})
