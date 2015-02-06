package property_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

var _ = Describe("Unfurl", func() {
	It("splits compound keys & injects empty maps", func() {
		furledProperties := Map{
			"a.e":   "value-e",
			"a.b.c": "value-c",
			"a.b.d": "value-d",
		}

		unfurledProperties, err := Unfurl(furledProperties)
		Expect(err).ToNot(HaveOccurred())
		Expect(unfurledProperties).To(Equal(Map{
			"a": Map{
				"b": Map{
					"c": "value-c",
					"d": "value-d",
				},
				"e": "value-e",
			},
		}))
	})

	It("allows map values", func() {
		partiallyFurledProperties := Map{
			"dns.db.database": "bosh",
			"dns.db.connection_options": Map{
				"max_connections": 32,
				"pool_timeout":    10,
			},
		}

		unfurledProperties, err := Unfurl(partiallyFurledProperties)
		Expect(err).ToNot(HaveOccurred())
		Expect(unfurledProperties).To(Equal(Map{
			"dns": Map{
				"db": Map{
					"database": "bosh",
					"connection_options": Map{
						"max_connections": 32,
						"pool_timeout":    10,
					},
				},
			},
		}))
	})

	It("errors if structure conflicts", func() {
		furledProperties := Map{
			"a.b":   "value-b", // b as string
			"a.b.c": "value-c", // b as map
		}

		_, err := Unfurl(furledProperties)
		Expect(err).To(HaveOccurred())
		// depending on map iteration order, either error type/message is possible
	})

	It("errors if structure conflicts with map value", func() {
		furledProperties := Map{
			"a.b.c": "value-c", // b as map
			"a": Map{
				"b": "value-b", // b as string
			},
		}

		_, err := Unfurl(furledProperties)
		Expect(err).To(HaveOccurred())
		// depending on map iteration order, either error type/message is possible
	})

	It("errors if keys are duplicated after unfurling", func() {
		furledProperties := Map{
			"a.b": "value-c", // b as string
			"a": Map{
				"b": "value-b", // b as string
			},
		}

		_, err := Unfurl(furledProperties)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Property collision unfurling"))
	})
})
