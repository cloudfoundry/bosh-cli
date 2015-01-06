package devicepathresolver

import (
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
)

type dummyDevicePathResolver struct{}

func NewDummyDevicePathResolver() dummyDevicePathResolver {
	return dummyDevicePathResolver{}
}

func (resolver dummyDevicePathResolver) GetRealDevicePath(diskSettings boshsettings.DiskSettings) (string, bool, error) {
	return diskSettings.Path, false, nil
}
