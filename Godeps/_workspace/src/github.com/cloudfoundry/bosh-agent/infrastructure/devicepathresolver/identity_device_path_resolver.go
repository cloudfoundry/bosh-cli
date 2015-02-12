package devicepathresolver

import (
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
)

type identityDevicePathResolver struct{}

func NewIdentityDevicePathResolver() identityDevicePathResolver {
	return identityDevicePathResolver{}
}

func (r identityDevicePathResolver) GetRealDevicePath(diskSettings boshsettings.DiskSettings) (string, bool, error) {
	return diskSettings.Path, false, nil
}
