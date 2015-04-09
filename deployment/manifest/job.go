package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type Job struct {
	Name               string
	Instances          int
	Lifecycle          JobLifecycle
	Templates          []ReleaseJobRef
	Networks           []JobNetwork
	PersistentDisk     int
	PersistentDiskPool string
	Properties         biproperty.Map
}

type JobLifecycle string

const (
	JobLifecycleService JobLifecycle = "service"
	JobLifecycleErrand  JobLifecycle = "errand"
)

type ReleaseJobRef struct {
	Name    string
	Release string
}

type JobNetwork struct {
	Name      string
	Default   []NetworkDefault
	StaticIPs []string
}

type NetworkDefault string

const (
	NetworkDefaultDNS     NetworkDefault = "dns"
	NetworkDefaultGateway NetworkDefault = "gateway"
)
