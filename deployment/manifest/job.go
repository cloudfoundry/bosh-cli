package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type Job struct {
	Name               string
	Instances          int
	Lifecycle          JobLifecycle
	Templates          []ReleaseJobRef
	Networks           []JobNetwork
	PersistentDisk     int
	PersistentDiskPool string
	Properties         bmproperty.Map
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
