package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type DiskPool struct {
	Name            string
	DiskSize        int
	CloudProperties bmproperty.Map
}
