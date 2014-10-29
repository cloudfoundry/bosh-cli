package templatescompiler

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
)

type jobEvaluationContext struct {
	relJob             bmrel.Job
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
	job bmrel.Job,
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
	convertedProperties, err := ec.convertForPropertyResolver(ec.relJob.Properties)
	if err != nil {
		return []byte{}, bosherr.WrapError(err, "Converting job properties for resolver")
	}

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

func (ec jobEvaluationContext) convertForPropertyResolver(properties map[string]bmrel.PropertyDefinition) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for propertyKey, property := range properties {
		defaultValue, err := property.Default()
		if err != nil {
			return result, bosherr.WrapError(err, "Retrieving default for property `%s'", propertyKey)
		}
		result[propertyKey] = defaultValue
	}

	return result, nil
}

func (ec jobEvaluationContext) buildNetworkContexts() map[string]networkContext {
	// IP is being returned by agent
	return map[string]networkContext{
		"default": networkContext{
			IP: "",
		},
	}
}
