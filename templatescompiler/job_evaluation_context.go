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
	releaseJob       bmreljob.Job
	jobProperties    bmproperty.Map
	globalProperties bmproperty.Map
	deploymentName   string
	logger           boshlog.Logger
	logTag           string
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index      int        `json:"index"`
	JobContext jobContext `json:"job"`
	Deployment string     `json:"deployment"`

	// Usually is accessed with <%= spec.networks.default.ip %>
	NetworkContexts map[string]networkContext `json:"networks"`

	//TODO: this should be a map[string]interface{}
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
	releaseJob bmreljob.Job,
	jobProperties bmproperty.Map,
	globalProperties bmproperty.Map,
	deploymentName string,
	logger boshlog.Logger,
) bmerbrenderer.TemplateEvaluationContext {
	return jobEvaluationContext{
		releaseJob:       releaseJob,
		jobProperties:    jobProperties,
		globalProperties: globalProperties,
		deploymentName:   deploymentName,
		logger:           logger,
		logTag:           "jobEvaluationContext",
	}
}

func (ec jobEvaluationContext) MarshalJSON() ([]byte, error) {
	defaultProperties := ec.propertyDefaults(ec.releaseJob.Properties)

	ec.logger.Debug(ec.logTag, "Original job '%s' property defaults: %#v", ec.releaseJob.Name, defaultProperties)

	properties, err := bmproperty.Unfurl(defaultProperties)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Unfurling job '%s' property defaults: %#v", ec.releaseJob.Name, defaultProperties)
	}
	ec.logger.Debug(ec.logTag, "Unfurled job '%s' property defaults: %#v", ec.releaseJob.Name, properties)

	ec.logger.Debug(ec.logTag, "Global properties: %#v", ec.globalProperties)
	err = bmproperty.Merge(properties, ec.globalProperties)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Merging global properties for job '%s'", ec.releaseJob.Name)
	}

	ec.logger.Debug(ec.logTag, "Job '%s' properties: %#v", ec.releaseJob.Name, ec.jobProperties)
	err = bmproperty.Merge(properties, ec.jobProperties)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Merging job properties for job '%s'", ec.releaseJob.Name)
	}

	ec.logger.Debug(ec.logTag, "Merged job '%s' properties: %#v", ec.releaseJob.Name, properties)

	context := RootContext{
		Index:           0,
		JobContext:      jobContext{Name: ec.releaseJob.Name},
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
