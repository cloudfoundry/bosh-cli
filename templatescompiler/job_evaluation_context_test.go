package templatescompiler_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"

	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

var _ = Describe("JobEvaluationContext", func() {
	var (
		generatedContext RootContext

		releaseJob              bmreljob.Job
		deploymentJobProperties bmproperty.Map
		deploymentProperties    bmproperty.Map
	)
	BeforeEach(func() {
		generatedContext = RootContext{}

		releaseJob = bmreljob.Job{
			Name: "fake-job-name",
			Properties: map[string]bmreljob.PropertyDefinition{
				"fake-default-property-first-level.fake-default-property-second-level": bmreljob.PropertyDefinition{
					Default: "default-property-value",
				},
			},
		}

		deploymentJobProperties = bmproperty.Map{
			"fake-job-property-first-level": bmproperty.Map{
				"fake-job-property-second-level": "job-property-value",
			},
		}

		deploymentProperties = bmproperty.Map{
			"fake-global-property-first-level": bmproperty.Map{
				"fake-global-property-second-level": "global-property-value",
			},
		}
	})

	JustBeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)

		jobEvaluationContext := NewJobEvaluationContext(
			releaseJob,
			deploymentJobProperties,
			deploymentProperties,
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

		Expect(generatedContext.Properties).To(Equal(bmproperty.Map{
			"fake-default-property-first-level": map[string]interface{}{
				"fake-default-property-second-level": "default-property-value",
			},
			"fake-job-property-first-level": map[string]interface{}{
				"fake-job-property-second-level": "job-property-value",
			},
			"fake-global-property-first-level": map[string]interface{}{
				"fake-global-property-second-level": "global-property-value",
			},
		}))
	})

	It("it has a network context section with empty IP", func() {
		Expect(generatedContext.NetworkContexts["default"].IP).To(Equal(""))
	})

	Context("when a job property overrides a global property", func() {
		BeforeEach(func() {
			deploymentJobProperties = bmproperty.Map{
				"fake-overridden-property-first-level": bmproperty.Map{
					"fake-overridden-property-second-level": "job-property-value",
				},
			}

			deploymentProperties = bmproperty.Map{
				"fake-overridden-property-first-level": bmproperty.Map{
					"fake-overridden-property-second-level": "global-property-value",
				},
			}
		})

		It("prefers job values over global values", func() {
			Expect(generatedContext.Properties).To(Equal(bmproperty.Map{
				"fake-default-property-first-level": map[string]interface{}{
					"fake-default-property-second-level": "default-property-value",
				},
				"fake-overridden-property-first-level": map[string]interface{}{
					"fake-overridden-property-second-level": "job-property-value",
				},
			}))
		})
	})

	Context("when a global property overrides a default property", func() {
		BeforeEach(func() {
			releaseJob.Properties = map[string]bmreljob.PropertyDefinition{
				"fake-property-first-level.fake-property-second-level": bmreljob.PropertyDefinition{
					Default: "default-property-value",
				},
			}

			deploymentJobProperties = bmproperty.Map{}

			deploymentProperties = bmproperty.Map{
				"fake-property-first-level": bmproperty.Map{
					"fake-property-second-level": "global-property-value",
				},
			}
		})

		It("prefers global values over default values", func() {
			Expect(generatedContext.Properties).To(Equal(bmproperty.Map{
				"fake-property-first-level": map[string]interface{}{
					"fake-property-second-level": "global-property-value",
				},
			}))
		})
	})

	Context("when a job property overrides a default property", func() {
		BeforeEach(func() {
			releaseJob.Properties = map[string]bmreljob.PropertyDefinition{
				"fake-property-first-level.fake-property-second-level": bmreljob.PropertyDefinition{
					Default: "default-property-value",
				},
			}

			deploymentJobProperties = bmproperty.Map{
				"fake-property-first-level": bmproperty.Map{
					"fake-property-second-level": "job-property-value",
				},
			}

			deploymentProperties = bmproperty.Map{}
		})

		It("prefers job values over default values", func() {
			Expect(generatedContext.Properties).To(Equal(bmproperty.Map{
				"fake-property-first-level": map[string]interface{}{
					"fake-property-second-level": "job-property-value",
				},
			}))
		})
	})
})
