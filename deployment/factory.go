package deployment

import (
	"time"

	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type Factory interface {
	NewDeployment(
		[]bminstance.Instance,
		[]bmdisk.Disk,
		[]bmstemcell.CloudStemcell,
	) Deployment
}

type factory struct {
	pingTimeout time.Duration
	pingDelay   time.Duration
}

func NewFactory(
	pingTimeout time.Duration,
	pingDelay time.Duration,
) Factory {
	return &factory{
		pingTimeout: pingTimeout,
		pingDelay:   pingDelay,
	}
}

func (f *factory) NewDeployment(
	instances []bminstance.Instance,
	disks []bmdisk.Disk,
	stemcells []bmstemcell.CloudStemcell,
) Deployment {
	return NewDeployment(
		instances,
		disks,
		stemcells,
		f.pingTimeout,
		f.pingDelay,
	)
}
