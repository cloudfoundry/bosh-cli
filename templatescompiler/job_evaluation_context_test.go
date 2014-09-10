package templatescompiler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

var _ = Describe("JobEvaluationContext", func() {
	var (
		generatedContext RootContext
	)
	BeforeEach(func() {
		job := bmrel.Job{
			Name: "fake-job-name",
			Properties: map[string]bmrel.PropertyDefinition{
				"first-level-prop.second-level-prop": bmrel.PropertyDefinition{
					Default: "fake-default",
				},
			},
		}

		manifestProperties := map[string]interface{}{
			"first-level-manifest-property": map[string]interface{}{
				"second-level-manifest-property": "manifest-property-value",
			},
		}
		logger := boshlog.NewLogger(boshlog.LevelNone)

		jobEvaluationContext := NewJobEvaluationContext(
			job,
			manifestProperties,
			"fake-deployment-name",
			logger,
		)

		generatedJSON, err := jobEvaluationContext.MarshalJSON()
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(generatedJSON, &generatedContext)
		Expect(err).ToNot(HaveOccurred())
	})

	It("generates correct json", func() {
		Expect(generatedContext.Index).To(Equal(0))
		Expect(generatedContext.JobContext.Name).To(Equal("fake-job-name"))
		Expect(generatedContext.Deployment).To(Equal("fake-deployment-name"))
		Expect(generatedContext.Properties["first-level-prop"]).To(Equal(
			map[string]interface{}{
				"second-level-prop": "fake-default",
			},
		))

		Expect(generatedContext.Properties["first-level-manifest-property"]).To(Equal(
			map[string]interface{}{
				"second-level-manifest-property": "manifest-property-value",
			},
		))
	})

	It("it has a network context section with empty IP", func() {
		Expect(generatedContext.NetworkContexts["default"].IP).To(Equal(""))
	})
})
