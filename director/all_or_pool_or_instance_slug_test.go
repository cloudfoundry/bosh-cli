package director_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/director"
)

var _ = Describe("NewAllOrPoolOrInstanceSlugFromString", func() {
	It("populates slug with empty name and index-or-id", func() {
		slug, err := NewAllOrPoolOrInstanceSlugFromString("")
		Expect(err).ToNot(HaveOccurred())
		Expect(slug.Name()).To(Equal(""))
		Expect(slug.IndexOrID()).To(Equal(""))
	})

	It("populates slug when name is just given", func() {
		slug, err := NewAllOrPoolOrInstanceSlugFromString("name")
		Expect(err).ToNot(HaveOccurred())
		Expect(slug.Name()).To(Equal("name"))
		Expect(slug.IndexOrID()).To(Equal(""))
	})

	It("populates slug when name and index-or-id is given", func() {
		slug, err := NewAllOrPoolOrInstanceSlugFromString("name/id")
		Expect(err).ToNot(HaveOccurred())
		Expect(slug.Name()).To(Equal("name"))
		Expect(slug.IndexOrID()).To(Equal("id"))
	})

	It("returns an error if string doesnt have 1 or 2 pieces", func() {
		_, err := NewAllOrPoolOrInstanceSlugFromString("1/2/3")
		Expect(err).To(Equal(errors.New("Expected pool or instance '1/2/3' to be in format 'name' or 'name/id-or-index'")))
	})

	It("returns an error if name is empty", func() {
		_, err := NewAllOrPoolOrInstanceSlugFromString("/")
		Expect(err).To(Equal(errors.New("Expected pool or instance '/' to specify non-empty name")))
	})

	It("returns an error if index-or-id is empty", func() {
		_, err := NewAllOrPoolOrInstanceSlugFromString("name/")
		Expect(err).To(Equal(errors.New("Expected instance 'name/' to specify non-empty ID or index")))
	})
})

var _ = Describe("AllPoolOrInstanceSlug", func() {
	Describe("InstanceSlug", func() {
		It("returns true and slug if name and id is set", func() {
			slug, ok := NewAllOrPoolOrInstanceSlug("name", "id").InstanceSlug()
			Expect(slug).To(Equal(NewInstanceSlug("name", "id")))
			Expect(ok).To(BeTrue())
		})

		It("returns false if name or id is not set", func() {
			slug, ok := NewAllOrPoolOrInstanceSlug("", "").InstanceSlug()
			Expect(slug).To(Equal(InstanceSlug{}))
			Expect(ok).To(BeFalse())

			slug, ok = NewAllOrPoolOrInstanceSlug("name", "").InstanceSlug()
			Expect(slug).To(Equal(InstanceSlug{}))
			Expect(ok).To(BeFalse())

			slug, ok = NewAllOrPoolOrInstanceSlug("", "id").InstanceSlug()
			Expect(slug).To(Equal(InstanceSlug{}))
			Expect(ok).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("returns empty if name or id is not set", func() {
			Expect(NewAllOrPoolOrInstanceSlug("", "").String()).To(Equal(""))
		})

		It("returns name string if id is not set", func() {
			Expect(NewAllOrPoolOrInstanceSlug("name", "").String()).To(Equal("name"))
		})

		It("returns name/id string if id is set", func() {
			Expect(NewAllOrPoolOrInstanceSlug("name", "id").String()).To(Equal("name/id"))
		})
	})

	Describe("UnmarshalFlag", func() {
		var (
			slug *AllOrPoolOrInstanceSlug
		)

		BeforeEach(func() {
			slug = &AllOrPoolOrInstanceSlug{}
		})

		It("populates slug with empty name and index-or-id", func() {
			err := slug.UnmarshalFlag("")
			Expect(err).ToNot(HaveOccurred())
			Expect(*slug).To(Equal(NewAllOrPoolOrInstanceSlug("", "")))
		})

		It("populates slug when name is just given", func() {
			err := slug.UnmarshalFlag("name")
			Expect(err).ToNot(HaveOccurred())
			Expect(*slug).To(Equal(NewAllOrPoolOrInstanceSlug("name", "")))
		})

		It("populates slug when name and index-or-id is given", func() {
			err := slug.UnmarshalFlag("name/id")
			Expect(err).ToNot(HaveOccurred())
			Expect(*slug).To(Equal(NewAllOrPoolOrInstanceSlug("name", "id")))
		})

		It("returns an error if string doesnt have 1 or 2 pieces", func() {
			err := slug.UnmarshalFlag("1/2/3")
			Expect(err).To(Equal(errors.New("Expected pool or instance '1/2/3' to be in format 'name' or 'name/id-or-index'")))
		})

		It("returns an error if name is empty", func() {
			err := slug.UnmarshalFlag("/")
			Expect(err).To(Equal(errors.New("Expected pool or instance '/' to specify non-empty name")))
		})

		It("returns an error if index-or-id is empty", func() {
			err := slug.UnmarshalFlag("name/")
			Expect(err).To(Equal(errors.New("Expected instance 'name/' to specify non-empty ID or index")))
		})
	})
})
