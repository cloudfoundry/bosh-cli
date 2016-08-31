package director_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/director"
)

var _ = Describe("SkipDrain", func() {
	Describe("AsQueryValue", func() {
		It("returns empty string when not skipping anything", func() {
			Expect(SkipDrain{}.AsQueryValue()).To(Equal(""))
		})

		It("returns '*' when skipping all", func() {
			Expect(SkipDrain{All: true}.AsQueryValue()).To(Equal("*"))
		})

		It("returns comma-separated slugs when skipping specific pools or instances", func() {
			skipDrain := SkipDrain{
				Slugs: []PoolOrInstanceSlug{
					NewPoolOrInstanceSlug("name1", ""),
					NewPoolOrInstanceSlug("name2", "id2"),
				},
			}
			Expect(skipDrain.AsQueryValue()).To(Equal("name1,name2/id2"))
		})
	})

	Describe("UnmarshalFlag", func() {
		var (
			skipDrain *SkipDrain
		)

		BeforeEach(func() {
			skipDrain = &SkipDrain{}
		})

		It("return skip drain for all when string is empty", func() {
			err := skipDrain.UnmarshalFlag("")
			Expect(err).ToNot(HaveOccurred())
			Expect(*skipDrain).To(Equal(SkipDrain{All: true}))
		})

		It("returns skip drain if slugs can be extracted", func() {
			err := skipDrain.UnmarshalFlag("name")
			Expect(err).ToNot(HaveOccurred())
			Expect(*skipDrain).To(Equal(SkipDrain{
				Slugs: []PoolOrInstanceSlug{
					NewPoolOrInstanceSlug("name", ""),
				},
			}))
		})

		It("returns skip drain if multiple slugs can be extracted", func() {
			err := skipDrain.UnmarshalFlag("name1,name2")
			Expect(err).ToNot(HaveOccurred())
			Expect(*skipDrain).To(Equal(SkipDrain{
				Slugs: []PoolOrInstanceSlug{
					NewPoolOrInstanceSlug("name1", ""),
					NewPoolOrInstanceSlug("name2", ""),
				},
			}))
		})

		It("returns skip drain if slugs can be extracted with ids", func() {
			err := skipDrain.UnmarshalFlag("name1/id1,name2")
			Expect(err).ToNot(HaveOccurred())
			Expect(*skipDrain).To(Equal(SkipDrain{
				Slugs: []PoolOrInstanceSlug{
					NewPoolOrInstanceSlug("name1", "id1"),
					NewPoolOrInstanceSlug("name2", ""),
				},
			}))
		})

		It("returns an error if slugs cannot be successfully extracted", func() {
			err := skipDrain.UnmarshalFlag("name/2/3")
			Expect(err).To(Equal(errors.New("Expected pool or instance 'name/2/3' to be in format 'name' or 'name/id-or-index'")))
		})
	})
})
