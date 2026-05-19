package manifest

import biproperty "github.com/cloudfoundry/bosh-utils/property"

type AvailabilityZone struct {
	Name            string
	CloudProperties biproperty.Map
}
