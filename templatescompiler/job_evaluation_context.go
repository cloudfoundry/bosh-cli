package templatescompiler

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/property"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	bierbrenderer "github.com/cloudfoundry/bosh-init/templatescompiler/erbrenderer"
)

type jobEvaluationContext struct {
	releaseJob       bireljob.Job
	jobProperties    biproperty.Map
	globalProperties biproperty.Map
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
	GlobalProperties  biproperty.Map `json:"global_properties"`  // values from manifest's top-level properties
	ClusterProperties biproperty.Map `json:"cluster_properties"` // values from manifest's jobs[].properties
	DefaultProperties biproperty.Map `json:"default_properties"` // values from release's job's spec
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
	releaseJob bireljob.Job,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	logger boshlog.Logger,
) bierbrenderer.TemplateEvaluationContext {
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

	context := RootContext{
		Index:             0,
		JobContext:        jobContext{Name: ec.releaseJob.Name},
		Deployment:        ec.deploymentName,
		NetworkContexts:   ec.buildNetworkContexts(),
		GlobalProperties:  ec.globalProperties,
		ClusterProperties: ec.jobProperties,
		DefaultProperties: defaultProperties,
	}

	ec.logger.Debug(ec.logTag, "Marshalling context %#v", context)

	jsonBytes, err := json.Marshal(context)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Marshalling job eval context: %#v", context)
	}

	return jsonBytes, nil
}

func (ec jobEvaluationContext) propertyDefaults(properties map[string]bireljob.PropertyDefinition) biproperty.Map {
	result := biproperty.Map{}
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
