package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("Job", func() {
	var job Job

	Describe("FindTemplateByValue", func() {
		Context("when a template with the value exists", func() {
			var expectedTemplate map[string]string

			BeforeEach(func() {
				expectedTemplate = map[string]string{
					"fake-template-name": "fake-template-value",
				}
				job = Job{
					Templates: expectedTemplate,
				}
			})

			It("returns the template and true", func() {
				actualTemplate, ok := job.FindTemplateByValue("fake-template-value")
				Expect(actualTemplate).To(Equal("fake-template-name"))
				Expect(ok).To(BeTrue())
			})
		})

		Context("when the template does not exist", func() {
			It("returns nil and false", func() {
				_, ok := job.FindTemplateByValue("nonsense")
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("property definition", func() {
		It("return correct primitive default value", func() {
			pd := PropertyDefinition{
				RawDefault: "a value",
			}
			Expect(pd.Default()).To(Equal("a value"))
		})

		It("returns correct values for nested default values", func() {
			pd := PropertyDefinition{
				RawDefault: map[interface{}]interface{}{
					"key": map[interface{}]interface{}{
						"nested-key": "value",
					},
				},
			}
			Expect(pd.Default()).To(Equal(map[string]interface{}{
				"key": map[string]interface{}{
					"nested-key": "value",
				},
			}))
		})
	})
})
