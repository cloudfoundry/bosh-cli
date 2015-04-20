package manifest

import (
	"strings"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type ResourcePool struct {
	Name            string
	Network         string
	CloudProperties biproperty.Map
	Env             biproperty.Map
	Stemcell        StemcellRef
}

type StemcellRef struct {
	URL string
}

func (s StemcellRef) Path() string {
	return strings.TrimPrefix(s.URL, "file://")
}
