package erbrenderer

import (
	"encoding/json"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type TemplateEvaluationContext interface {
	MarshalJSON() ([]byte, error)
}

type templateEvaluationContext struct {
	relJob             bmreljob.Job
	manifestProperties map[string]interface{}
	deploymentName     string
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index int `json:"index"`

	JobContext jobContext `json:"job"`

	Deployment string `json:"deployment"`

	Properties map[string]interface{} `json:"properties"`
}

type jobContext struct {
	Name string `json:"name"`
}

func NewTemplateEvaluationContext(
	job bmreljob.Job,
	manifestProperties map[string]interface{},
	deploymentName string,
) TemplateEvaluationContext {
	return templateEvaluationContext{
		relJob:             job,
		manifestProperties: manifestProperties,
		deploymentName:     deploymentName,
	}
}

func (t templateEvaluationContext) MarshalJSON() ([]byte, error) {
	properties := NewPropertiesResolver(t.relJob.Properties, t.manifestProperties).Resolve()

	context := RootContext{
		Index:      0,
		JobContext: jobContext{Name: t.relJob.Name},
		Deployment: t.deploymentName,

		Properties: properties,
	}

	return json.Marshal(context)
}
