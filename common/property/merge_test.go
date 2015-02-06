package property_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

var _ = Describe("Merge", func() {
	It("copies values from the right onto the left", func() {
		leftMap := Map{
			"a": Map{
				"b": Map{
					"c": "value-c-left",
					"d": "value-d",
				},
			},
		}
		rightMap := Map{
			"a": Map{
				"b": Map{
					"c": "value-c-right",
					"e": "value-e",
				},
			},
			"x": "value-x",
		}

		err := Merge(leftMap, rightMap)
		Expect(err).ToNot(HaveOccurred())
		Expect(leftMap).To(Equal(Map{
			"a": Map{
				"b": Map{
					"c": "value-c-right",
					"d": "value-d",
					"e": "value-e",
				},
			},
			"x": "value-x",
		}))
	})
})
