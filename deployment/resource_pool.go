package deployment

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type ResourcePool struct {
	Name               string                      `yaml:"name"`
	Network            string                      `yaml:"network"`
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
	RawEnv             map[interface{}]interface{} `yaml:"env"`
}

func (rp ResourcePool) Env() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(rp.RawEnv)
}

func (rp ResourcePool) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(rp.RawCloudProperties)
}
