package deployment

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type ReleaseJobRef struct {
	Name    string
	Release string
}

type Job struct {
	Name      string
	Instances int
	Templates []ReleaseJobRef
}

type Deployment struct {
	Name          string
	Properties    map[string]interface{}
	Jobs          []Job
	Networks      []Network
	ResourcePools []ResourcePool
}

func (d Deployment) NetworksSpec() (map[string]interface{}, error) {
	result := map[string]interface{}{}

	for _, network := range d.Networks {
		spec, err := network.Spec()
		if err != nil {
			return result, bosherr.WrapError(err, "Building networksspec")
		}
		result[network.Name] = spec[network.Name]
	}

	return result, nil
}
