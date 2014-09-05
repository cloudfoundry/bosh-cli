package templatescompiler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"

	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

var _ = Describe("JobEvaluationContext", func() {
	var jobEvaluationContext bmerbrenderer.TemplateEvaluationContext
	BeforeEach(func() {
		job := bmreljob.Job{
			Name: "fake-job-name",
			Properties: map[string]bmreljob.PropertyDefinition{
				"first-level-prop.second-level-prop": bmreljob.PropertyDefinition{
					Default: "fake-default",
				},
			},
		}

		manifestProperties := map[string]interface{}{}

		jobEvaluationContext = NewJobEvaluationContext(
			job,
			manifestProperties,
			"fake-deployment-name",
		)
	})

	It("generates correct json", func() {
		generatedJSON, err := jobEvaluationContext.MarshalJSON()
		Expect(err).ToNot(HaveOccurred())

		var generatedContext RootContext
		err = json.Unmarshal(generatedJSON, &generatedContext)
		Expect(err).ToNot(HaveOccurred())

		Expect(generatedContext.Index).To(Equal(0))
		Expect(generatedContext.JobContext.Name).To(Equal("fake-job-name"))
		Expect(generatedContext.Deployment).To(Equal("fake-deployment-name"))
		Expect(generatedContext.Properties).To(Equal(map[string]interface{}{
			"first-level-prop": map[string]interface{}{
				"second-level-prop": "fake-default",
			},
		}))
	})
})
