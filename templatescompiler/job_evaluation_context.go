package templatescompiler

import (
	"encoding/json"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type jobEvaluationContext struct {
	relJob             bmreljob.Job
	manifestProperties map[string]interface{}
	deploymentName     string
	logger             boshlog.Logger
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index      int        `json:"index"`
	JobContext jobContext `json:"job"`
	Deployment string     `json:"deployment"`

	// Usually is accessed with <%= spec.networks.default.ip %>
	NetworkContexts map[string]networkContext `json:"networks"`

	Properties map[string]interface{} `json:"properties"`
}

type jobContext struct {
	Name string `json:"name"`
}

type networkContext struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

const logTag = "JobEvaluationContext"

func NewJobEvaluationContext(
	job bmreljob.Job,
	manifestProperties map[string]interface{},
	deploymentName string,
	logger boshlog.Logger,
) bmerbrenderer.TemplateEvaluationContext {
	return jobEvaluationContext{
		relJob:             job,
		manifestProperties: manifestProperties,
		deploymentName:     deploymentName,
		logger:             logger,
	}
}

func (ec jobEvaluationContext) MarshalJSON() ([]byte, error) {
	convertedProperties := ec.convertForPropertyResolver(ec.relJob.Properties)
	properties := bmerbrenderer.NewPropertiesResolver(convertedProperties, ec.manifestProperties).Resolve()

	context := RootContext{
		Index:           0,
		JobContext:      jobContext{Name: ec.relJob.Name},
		Deployment:      ec.deploymentName,
		NetworkContexts: ec.buildNetworkContexts(),
		Properties:      properties,
	}

	ec.logger.Debug(logTag, "Marshalling context %#v", context)

	return json.Marshal(context)
}

func (ec jobEvaluationContext) convertForPropertyResolver(properties map[string]bmreljob.PropertyDefinition) map[string]interface{} {
	result := map[string]interface{}{}
	for propertyKey, property := range properties {
		result[propertyKey] = property.Default
	}

	return result
}

func (ec jobEvaluationContext) buildNetworkContexts() map[string]networkContext {
	// IP is being returned by agent
	return map[string]networkContext{
		"default": networkContext{
			IP: "",
		},
	}
}
