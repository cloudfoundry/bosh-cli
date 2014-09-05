package templatescompiler

import (
	"encoding/json"

	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type jobEvaluationContext struct {
	relJob             bmreljob.Job
	manifestProperties map[string]interface{}
	deploymentName     string
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index      int                    `json:"index"`
	JobContext jobContext             `json:"job"`
	Deployment string                 `json:"deployment"`
	Properties map[string]interface{} `json:"properties"`
}

type jobContext struct {
	Name string `json:"name"`
}

func NewJobEvaluationContext(
	job bmreljob.Job,
	manifestProperties map[string]interface{},
	deploymentName string,
) bmerbrenderer.TemplateEvaluationContext {
	return jobEvaluationContext{
		relJob:             job,
		manifestProperties: manifestProperties,
		deploymentName:     deploymentName,
	}
}

func (t jobEvaluationContext) MarshalJSON() ([]byte, error) {
	convertedProperties := t.convertForPropertyResolver(t.relJob.Properties)
	properties := bmerbrenderer.NewPropertiesResolver(convertedProperties, t.manifestProperties).Resolve()

	context := RootContext{
		Index:      0,
		JobContext: jobContext{Name: t.relJob.Name},
		Deployment: t.deploymentName,
		Properties: properties,
	}

	return json.Marshal(context)
}

func (t jobEvaluationContext) convertForPropertyResolver(properties map[string]bmreljob.PropertyDefinition) map[string]interface{} {
	result := map[string]interface{}{}
	for propertyKey, property := range properties {
		result[propertyKey] = property.Default
	}

	return result
}
