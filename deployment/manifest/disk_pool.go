package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type DiskPool struct {
	Name               string                      `yaml:"name"`
	DiskSize           int                         `yaml:"disk_size"`
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

func (dp DiskPool) CloudProperties() (bmproperty.Map, error) {
	return bmproperty.BuildMap(dp.RawCloudProperties)
}
