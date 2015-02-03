package manifest

import (
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type DiskPool struct {
	Name            string
	DiskSize        int
	CloudProperties bmproperty.Map
}
