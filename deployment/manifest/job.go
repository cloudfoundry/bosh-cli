package manifest

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
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

func (j *Job) ReleaseJobReferences() []bmreljob.Reference {
	jobRefs := make([]bmreljob.Reference, len(j.Templates), len(j.Templates))
	for i, jobRef := range j.Templates {
		jobRefs[i] = bmreljob.Reference{
			Name:    jobRef.Name,
			Release: jobRef.Release,
		}
	}
	return jobRefs
}
