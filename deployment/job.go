package deployment

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type Job struct {
	Name               string
	Instances          int
	Lifecycle          JobLifecycle
	Templates          []ReleaseJobRef
	Networks           []JobNetwork
	PersistentDisk     int                         `yaml:"persistent_disk"`
	PersistentDiskPool string                      `yaml:"persistent_disk_pool"`
	RawProperties      map[interface{}]interface{} `yaml:"properties"`
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
	StaticIPs []string `yaml:"static_ips"`
}

type NetworkDefault string

const (
	NetworkDefaultDNS     NetworkDefault = "dns"
	NetworkDefaultGateway NetworkDefault = "gateway"
)

func (j *Job) Properties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(j.RawProperties)
}

