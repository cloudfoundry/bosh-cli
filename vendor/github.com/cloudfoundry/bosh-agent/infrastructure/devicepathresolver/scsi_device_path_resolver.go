package devicepathresolver

import (
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type scsiDevicePathResolver struct {
	scsiVolumeIDPathResolver DevicePathResolver
	scsiIDPathResolver       DevicePathResolver
}

func NewScsiDevicePathResolver(
	scsiVolumeIDPathResolver DevicePathResolver,
	scsiIDPathResolver DevicePathResolver,
) DevicePathResolver {
	return scsiDevicePathResolver{
		scsiVolumeIDPathResolver: scsiVolumeIDPathResolver,
		scsiIDPathResolver:       scsiIDPathResolver,
	}
}

func (sr scsiDevicePathResolver) GetRealDevicePath(diskSettings boshsettings.DiskSettings) (string, bool, error) {
	if len(diskSettings.DeviceID) > 0 {
		return sr.scsiIDPathResolver.GetRealDevicePath(diskSettings)
	}

	if len(diskSettings.VolumeID) > 0 {
		return sr.scsiVolumeIDPathResolver.GetRealDevicePath(diskSettings)
	}

	return "", false, bosherr.Error("Neither ID nor VolumeID provided in disk settings")
}
