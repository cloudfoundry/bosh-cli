package keystringifier_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

var _ = Describe("KeyStringifier", func() {
	var keyStringifier KeyStringifier

	BeforeEach(func() {
		keyStringifier = NewKeyStringifier()
	})

	Describe("ConvertMap", func() {
		It("returns a map with stringified keys", func() {
			input := map[interface{}]interface{}{
				"fake-key": "fake-value",
			}

			result, err := keyStringifier.ConvertMap(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(map[string]interface{}{
				"fake-key": "fake-value",
			}))
		})

		It("returns a map with nested stringified keys", func() {
			input := map[interface{}]interface{}{
				"fake-key": map[interface{}]interface{}{
					"nested-key": "fake-value",
				},
			}

			result, err := keyStringifier.ConvertMap(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(map[string]interface{}{
				"fake-key": map[string]interface{}{
					"nested-key": "fake-value",
				},
			}))
		})

		Context("when map contains non-string key", func() {
			It("returns an error", func() {
				input := map[interface{}]interface{}{
					123: "fake-value",
				}

				_, err := keyStringifier.ConvertMap(input)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Map contains non-string key"))
			})
		})
	})
})
