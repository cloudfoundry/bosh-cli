package devicepathresolver

import (
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
)

type identityDevicePathResolver struct{}

func NewIdentityDevicePathResolver() DevicePathResolver {
	return identityDevicePathResolver{}
}

func (r identityDevicePathResolver) GetRealDevicePath(diskSettings boshsettings.DiskSettings) (string, bool, error) {
	return diskSettings.Path, false, nil
}
