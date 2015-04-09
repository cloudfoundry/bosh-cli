package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type DiskPool struct {
	Name            string
	DiskSize        int
	CloudProperties biproperty.Map
}
