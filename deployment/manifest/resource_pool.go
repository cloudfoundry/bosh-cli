package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type ResourcePool struct {
	Name            string
	Network         string
	CloudProperties biproperty.Map
	Env             biproperty.Map
}
