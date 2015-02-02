package erbrenderer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

var _ = Describe("PropertiesResolver", func() {
	var (
		propertiesResolver PropertiesResolver
		defaults           bmproperty.Map
		values             bmproperty.Map
	)

	Context("when value is specified for nested property", func() {
		BeforeEach(func() {
			values = bmproperty.Map{
				"first-level-prop": bmproperty.Map{
					"second-level-prop": "original-value",
				},
			}
			defaults = bmproperty.Map{
				"first-level-prop.second-level-prop": "default-value",
			}

			propertiesResolver = NewPropertiesResolver(defaults, values)
		})

		It("returns the specified value", func() {
			properties := propertiesResolver.Resolve()
			Expect(properties).To(Equal(bmproperty.Map{
				"first-level-prop": bmproperty.Map{
					"second-level-prop": "original-value",
				},
			}))
		})
	})

	Context("when value is not specified for nested property", func() {
		BeforeEach(func() {
			values = bmproperty.Map{}
		})

		Context("when default property is specified", func() {
			BeforeEach(func() {
				defaults = bmproperty.Map{
					"first-level-prop.second-level-prop": "default-value",
				}

				propertiesResolver = NewPropertiesResolver(defaults, values)
			})

			It("uses default property", func() {
				properties := propertiesResolver.Resolve()
				Expect(properties).To(Equal(bmproperty.Map{
					"first-level-prop": bmproperty.Map{
						"second-level-prop": "default-value",
					},
				}))
			})
		})

		Context("when default property is nil", func() {
			BeforeEach(func() {
				defaults = bmproperty.Map{
					"first-level-prop.second-level-prop": nil,
				}

				propertiesResolver = NewPropertiesResolver(defaults, values)
			})

			It("uses default property", func() {
				properties := propertiesResolver.Resolve()
				Expect(properties).To(Equal(bmproperty.Map{
					"first-level-prop": bmproperty.Map{
						"second-level-prop": nil,
					},
				}))
			})
		})

		Context("when default property is empty string", func() {
			BeforeEach(func() {
				defaults = bmproperty.Map{
					"first-level-prop.second-level-prop": "",
				}

				propertiesResolver = NewPropertiesResolver(defaults, values)
			})

			It("uses default property", func() {
				properties := propertiesResolver.Resolve()
				Expect(properties).To(Equal(bmproperty.Map{
					"first-level-prop": bmproperty.Map{
						"second-level-prop": "",
					},
				}))
			})
		})
	})
})
