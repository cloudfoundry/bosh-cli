package templatescompiler

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	bierbrenderer "github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
)

// LinkInstanceSpec matches the per-instance fields whitelisted by the BOSH director's
// instance_spec.rb when serialising link data for ERB rendering.
type LinkInstanceSpec struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Index     int    `json:"index"`
	Bootstrap bool   `json:"bootstrap"`
	AZ        string `json:"az"`
	Address   string `json:"address"`
}

// LinkSpec is the full link object exposed to ERB templates.
// Field names match the whitelisted keys from the BOSH director's instance_spec.rb:
//
//	address, default_network, deployment_name, domain, group_name,
//	instance_group, instances, properties, use_link_dns_names, use_short_dns_addresses
//
// The optional Address field is only set for manual links; its presence signals
// erb_renderer.rb to use ManualLinkDnsEncoder so that link.address() returns the
// fixed address string instead of raising NotImplementedError.
type LinkSpec struct {
	DeploymentName       string                 `json:"deployment_name"`
	Domain               string                 `json:"domain"`
	InstanceGroup        string                 `json:"instance_group"`
	DefaultNetwork       string                 `json:"default_network"`
	GroupName            string                 `json:"group_name"`
	Instances            []LinkInstanceSpec     `json:"instances"`
	Properties           map[string]interface{} `json:"properties"`
	UseLinkDNSNames      bool                   `json:"use_link_dns_names"`
	UseShortDNSAddresses bool                   `json:"use_short_dns_addresses"`
	// Address is only set for manual links.
	Address string `json:"address,omitempty"`
}

// InstanceSpec carries per-instance data exposed to ERB templates through the spec object.
// It should be built by the state builder for each instance being rendered.
type InstanceSpec struct {
	// Name is the instance group name (spec.name and spec.job.name).
	Name string
	// Index is the ordinal instance index within the group (spec.index).
	Index int
	// AZ is the availability zone the instance is placed in (spec.az).
	AZ string
	// Bootstrap is true when this is the first instance of the group (spec.bootstrap).
	Bootstrap bool
	// Address is the default network address, used for spec.address and spec.ip.
	Address string
	// Networks maps network name → per-network context (spec.networks.<name>).
	Networks map[string]NetworkSpecContext
	// PersistentDisk is the persistent disk size in MB; 0 means no disk (spec.persistent_disk).
	PersistentDisk int
	// ReleaseNamesByJob maps release job name → release name (for spec.release.name lookup).
	ReleaseNamesByJob map[string]string
	// Links holds the resolved links for every release job in this instance group,
	// keyed first by job template name then by link name.
	// This mirrors the director's namespace_links_to_current_job pattern.
	Links map[string]map[string]LinkSpec
}

// NetworkSpecContext holds per-network data exposed through spec.networks.<name>.
type NetworkSpecContext struct {
	IP      string   `json:"ip"`
	Netmask string   `json:"netmask"`
	Gateway string   `json:"gateway"`
	DNS     []string `json:"dns,omitempty"`
}

type jobEvaluationContext struct {
	releaseJob           bireljob.Job
	releaseJobProperties *biproperty.Map
	jobProperties        biproperty.Map
	globalProperties     biproperty.Map
	deploymentName       string
	spec                 InstanceSpec
	uuidGen              boshuuid.Generator
	logger               boshlog.Logger
	logTag               string
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index     int        `json:"index"`
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	AZ        string     `json:"az"`
	Bootstrap bool       `json:"bootstrap"`
	// IP is the IP address of the default network (spec.ip).
	IP         string     `json:"ip"`
	Address    string     `json:"address,omitempty"`
	JobContext jobContext `json:"job"`
	Deployment string     `json:"deployment"`

	// Usually is accessed with <%= spec.networks.default.ip %>
	NetworkContexts map[string]NetworkSpecContext `json:"networks"`

	// PersistentDisk is 0 when no persistent disk is configured (spec.persistent_disk).
	PersistentDisk int `json:"persistent_disk"`
	// DnsDomainName is the configured root domain name for the Director (spec.dns_domain_name).
	DnsDomainName string `json:"dns_domain_name"`
	// ReleaseContext exposes spec.release.name and spec.release.version.
	ReleaseContext releaseContext `json:"release"`

	//TODO: this should be a map[string]interface{}
	GlobalProperties  biproperty.Map  `json:"global_properties"`  // values from manifest's top-level properties
	ClusterProperties biproperty.Map  `json:"cluster_properties"` // values from instance group (deployment job) properties
	JobProperties     *biproperty.Map `json:"job_properties"`     // values from release job (aka template) properties
	DefaultProperties biproperty.Map  `json:"default_properties"` // values from release's job's spec

	// JobTemplateName is the name of the release job being rendered.
	// erb_renderer.rb uses it to select the correct slice of Links.
	JobTemplateName string `json:"job_template_name"`
	// Links holds the resolved links for all release jobs in this instance group,
	// keyed by job template name then by link name. This matches the BOSH director's
	// namespace_links_to_current_job pattern.
	Links map[string]map[string]LinkSpec `json:"links"`
}

type jobContext struct {
	Name string `json:"name"`
}

type releaseContext struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewJobEvaluationContext(
	releaseJob bireljob.Job,
	releaseJobProperties *biproperty.Map,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	spec InstanceSpec,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) bierbrenderer.TemplateEvaluationContext {
	return jobEvaluationContext{
		releaseJob:           releaseJob,
		releaseJobProperties: releaseJobProperties,
		jobProperties:        jobProperties,
		globalProperties:     globalProperties,
		deploymentName:       deploymentName,
		spec:                 spec,
		uuidGen:              uuidGen,
		logTag:               "jobEvaluationContext",
		logger:               logger,
	}
}

func (ec jobEvaluationContext) MarshalJSON() ([]byte, error) {
	defaultProperties := ec.propertyDefaults(ec.releaseJob.Properties)
	var err error

	networks := ec.spec.Networks
	if networks == nil {
		networks = map[string]NetworkSpecContext{}
	}

	releaseName := ec.spec.ReleaseNamesByJob[ec.releaseJob.Name()]

	links := ec.spec.Links
	if links == nil {
		links = map[string]map[string]LinkSpec{}
	}

	context := RootContext{
		Index:             ec.spec.Index,
		AZ:                ec.spec.AZ,
		Name:              ec.spec.Name,
		Bootstrap:         ec.spec.Bootstrap,
		IP:                ec.spec.Address,
		JobContext:        jobContext{Name: ec.spec.Name},
		Deployment:        ec.deploymentName,
		NetworkContexts:   networks,
		PersistentDisk:    ec.spec.PersistentDisk,
		DnsDomainName:     "bosh",
		ReleaseContext:    releaseContext{Name: releaseName, Version: ec.releaseJob.Fingerprint()},
		GlobalProperties:  ec.globalProperties,
		ClusterProperties: ec.jobProperties,
		JobProperties:     ec.releaseJobProperties,
		DefaultProperties: defaultProperties,
		JobTemplateName:   ec.releaseJob.Name(),
		Links:             links,
	}

	if ec.spec.Address != "" {
		context.Address = ec.spec.Address
	}

	context.ID, err = ec.uuidGen.Generate()
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Setting job eval context's ID to UUID: %#v", context)
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
