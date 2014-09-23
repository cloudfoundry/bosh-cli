package cdrom

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshdevutil "github.com/cloudfoundry/bosh-agent/platform/deviceutil"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type cdUtil struct {
	settingsMountPath string
	fs                boshsys.FileSystem
	cdrom             Cdrom
}

func NewCdUtil(settingsMountPath string, fs boshsys.FileSystem, cdrom Cdrom) boshdevutil.DeviceUtil {
	return cdUtil{
		settingsMountPath: settingsMountPath,
		fs:                fs,
		cdrom:             cdrom,
	}
}

func (util cdUtil) GetFilesContents(fileNames []string) ([][]byte, error) {
	err := util.cdrom.WaitForMedia()
	if err != nil {
		return [][]byte{}, bosherr.WrapError(err, "Waiting for CDROM to be ready")
	}

	err = util.fs.MkdirAll(util.settingsMountPath, os.FileMode(0700))
	if err != nil {
		return [][]byte{}, bosherr.WrapError(err, "Creating CDROM mount point")
	}

	err = util.cdrom.Mount(util.settingsMountPath)
	if err != nil {
		return [][]byte{}, bosherr.WrapError(err, "Mounting CDROM")
	}

	contents := [][]byte{}
	for _, fileName := range fileNames {
		settingsPath := filepath.Join(util.settingsMountPath, fileName)
		stringContents, err := util.fs.ReadFile(settingsPath)
		if err != nil {
			return [][]byte{}, bosherr.WrapError(err, "Reading from CDROM")
		}

		contents = append(contents, []byte(stringContents))
	}

	err = util.cdrom.Unmount()
	if err != nil {
		return [][]byte{}, bosherr.WrapError(err, "Unmounting CDROM")
	}

	err = util.cdrom.Eject()
	if err != nil {
		return [][]byte{}, bosherr.WrapError(err, "Ejecting CDROM")
	}

	return contents, nil
}
