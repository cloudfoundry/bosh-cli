package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type ResourcePool struct {
	Name               string                      `yaml:"name"`
	Network            string                      `yaml:"network"`
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
	RawEnv             map[interface{}]interface{} `yaml:"env"`
}

func (rp ResourcePool) Env() (bmproperty.Map, error) {
	return bmproperty.BuildMap(rp.RawEnv)
}

func (rp ResourcePool) CloudProperties() (bmproperty.Map, error) {
	return bmproperty.BuildMap(rp.RawCloudProperties)
}
