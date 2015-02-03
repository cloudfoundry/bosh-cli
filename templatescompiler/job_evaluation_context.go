package templatescompiler

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
)

type jobEvaluationContext struct {
	relJob             bmreljob.Job
	manifestProperties bmproperty.Map
	deploymentName     string
	logger             boshlog.Logger
	logTag             string
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index      int        `json:"index"`
	JobContext jobContext `json:"job"`
	Deployment string     `json:"deployment"`

	// Usually is accessed with <%= spec.networks.default.ip %>
	NetworkContexts map[string]networkContext `json:"networks"`

	Properties bmproperty.Map `json:"properties"`
}

type jobContext struct {
	Name string `json:"name"`
}

type networkContext struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

func NewJobEvaluationContext(
	job bmreljob.Job,
	manifestProperties bmproperty.Map,
	deploymentName string,
	logger boshlog.Logger,
) bmerbrenderer.TemplateEvaluationContext {
	return jobEvaluationContext{
		relJob:             job,
		manifestProperties: manifestProperties,
		deploymentName:     deploymentName,
		logger:             logger,
		logTag:             "jobEvaluationContext",
	}
}

func (ec jobEvaluationContext) MarshalJSON() ([]byte, error) {
	propertyDefaults := ec.propertyDefaults(ec.relJob.Properties)

	ec.logger.Debug(ec.logTag, "Job '%s' properties: %#v", ec.relJob.Name, propertyDefaults)
	ec.logger.Debug(ec.logTag, "Deployment manifest properties: %#v", ec.manifestProperties)

	properties := bmerbrenderer.NewPropertiesResolver(propertyDefaults, ec.manifestProperties).Resolve()

	ec.logger.Debug(ec.logTag, "Resolved Job '%s' properties: %#v", ec.relJob.Name, properties)

	context := RootContext{
		Index:           0,
		JobContext:      jobContext{Name: ec.relJob.Name},
		Deployment:      ec.deploymentName,
		NetworkContexts: ec.buildNetworkContexts(),
		Properties:      properties,
	}

	ec.logger.Debug(ec.logTag, "Marshalling context %#v", context)

	jsonBytes, err := json.Marshal(context)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Marshalling job eval context: %#v", context)
	}

	return jsonBytes, nil
}

func (ec jobEvaluationContext) propertyDefaults(properties map[string]bmreljob.PropertyDefinition) bmproperty.Map {
	result := bmproperty.Map{}
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
