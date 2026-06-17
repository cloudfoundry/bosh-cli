package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Job struct {
	Name               string
	Instances          int
	Lifecycle          JobLifecycle
	Templates          []ReleaseJobRef
	Networks           []JobNetwork
	PersistentDisk     int
	PersistentDiskPool string
	ResourcePool       string
	Properties         biproperty.Map
	AZs                []string
}

type JobLifecycle string

const (
	JobLifecycleService JobLifecycle = "service"
	JobLifecycleErrand  JobLifecycle = "errand"
)

type ReleaseJobRef struct {
	Name       string
	Release    string
	Properties *biproperty.Map
	// Consumes is keyed by the link name declared in the job's spec.
	// A nil entry means the link is explicitly disabled ("nil" in YAML).
	// A non-nil entry with IsBlocked=true means the link was set to nil in YAML.
	Consumes map[string]ManifestConsumesEntry
}

// ManifestConsumesEntry captures the per-link override from the deployment manifest.
// Exactly one of the three modes applies:
//  1. IsBlocked=true   → link is explicitly disabled (consumes: {db: nil})
//  2. IsManual=true    → at least one of Instances/Properties/Address is set
//  3. From != ""       → redirect to a different provider by name
//
// Modes 2 and 3 are mutually exclusive in practice but the resolver handles them in order.
type ManifestConsumesEntry struct {
	IsBlocked bool
	IsManual  bool
	// Manual fields (MANUAL_LINK_KEYS: instances, properties, address)
	Instances  []ManualLinkInstance
	Properties map[string]interface{}
	Address    string
	// Alias/redirect
	From string
}

// ManualLinkInstance is a single entry in a manual link's instances list.
type ManualLinkInstance struct {
	Address string
}

type JobNetwork struct {
	Name      string
	Defaults  []NetworkDefault
	StaticIPs []string
}

type NetworkDefault string

const (
	NetworkDefaultDNS     NetworkDefault = "dns"
	NetworkDefaultGateway NetworkDefault = "gateway"
)
