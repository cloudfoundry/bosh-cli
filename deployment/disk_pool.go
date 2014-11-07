package deployment

import (
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type DiskPool struct {
	Name               string                      `yaml:"name"`
	Size               int                         `yaml:"disk_size"`
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

func (dp DiskPool) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(dp.RawCloudProperties)
}
