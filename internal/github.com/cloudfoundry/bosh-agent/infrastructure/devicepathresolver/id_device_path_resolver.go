package devicepathresolver

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	boshudev "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/udevdevice"
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
)

type idDevicePathResolver struct {
	diskWaitTimeout time.Duration
	udev            boshudev.UdevDevice
	fs              boshsys.FileSystem
}

func NewIDDevicePathResolver(
	diskWaitTimeout time.Duration,
	udev boshudev.UdevDevice,
	fs boshsys.FileSystem,
) DevicePathResolver {
	return idDevicePathResolver{
		diskWaitTimeout: diskWaitTimeout,
		udev:            udev,
		fs:              fs,
	}
}

func (idpr idDevicePathResolver) GetRealDevicePath(diskSettings boshsettings.DiskSettings) (string, bool, error) {
	if diskSettings.ID == "" {
		return "", false, bosherr.Errorf("Disk ID is not set")
	}

	if len(diskSettings.ID) < 20 {
		return "", false, bosherr.Errorf("Disk ID is not the correct format")
	}

	err := idpr.udev.Trigger()
	if err != nil {
		return "", false, bosherr.WrapError(err, "Running udevadm trigger")
	}

	err = idpr.udev.Settle()
	if err != nil {
		return "", false, bosherr.WrapError(err, "Running udevadm settle")
	}

	stopAfter := time.Now().Add(idpr.diskWaitTimeout)
	found := false

	var realPath string

	diskID := diskSettings.ID[0:20]

	for !found {
		if time.Now().After(stopAfter) {
			return "", true, bosherr.Errorf("Timed out getting real device path for '%s'", diskID)
		}

		time.Sleep(100 * time.Millisecond)

		deviceIDPath := filepath.Join(string(os.PathSeparator), "dev", "disk", "by-id", fmt.Sprintf("virtio-%s", diskID))

		realPath, err = idpr.fs.ReadLink(deviceIDPath)
		if err != nil {
			continue
		}

		if idpr.fs.FileExists(realPath) {
			found = true
		}
	}

	return realPath, false, nil
}
