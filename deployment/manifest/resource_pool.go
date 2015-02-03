package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type ResourcePool struct {
	Name            string
	Network         string
	CloudProperties bmproperty.Map
	Env             bmproperty.Map
}
