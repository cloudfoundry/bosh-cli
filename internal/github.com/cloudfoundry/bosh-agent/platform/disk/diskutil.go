package disk

import (
	"path/filepath"

	boshdevutil "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/deviceutil"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
)

type diskUtil struct {
	diskPath string
	mounter  Mounter
	fs       boshsys.FileSystem

	logTag string
	logger boshlog.Logger
}

func NewDiskUtil(diskPath string, mounter Mounter, fs boshsys.FileSystem, logger boshlog.Logger) boshdevutil.DeviceUtil {
	return diskUtil{
		diskPath: diskPath,
		mounter:  mounter,
		fs:       fs,

		logTag: "diskUtil",
		logger: logger,
	}
}

func (util diskUtil) GetFilesContents(fileNames []string) ([][]byte, error) {
	if !util.fs.FileExists(util.diskPath) {
		return [][]byte{}, bosherr.Errorf("Failed to get file contents, disk path '%s' does not exist", util.diskPath)
	}

	tempDir, err := util.fs.TempDir("diskutil")
	if err != nil {
		return [][]byte{}, bosherr.WrapError(err, "Creating temporary disk mount point")
	}

	defer util.fs.RemoveAll(tempDir)

	err = util.mounter.Mount(util.diskPath, tempDir)
	if err != nil {
		return [][]byte{}, bosherr.WrapErrorf(err, "Mounting disk path '%s' to '%s'", util.diskPath, tempDir)
	}

	util.logger.Debug(util.logTag, "Mounted disk path '%s' to '%s'", util.diskPath, tempDir)

	contents := [][]byte{}

	for _, fileName := range fileNames {
		diskFilePath := filepath.Join(tempDir, fileName)

		util.logger.Debug(util.logTag, "Reading contents of '%s'", diskFilePath)

		content, err := util.fs.ReadFile(diskFilePath)
		if err != nil {
			// todo unmount before removing
			util.unmount(tempDir)
			return [][]byte{}, bosherr.WrapErrorf(err, "Reading from disk file '%s'", diskFilePath)
		}

		util.logger.Debug(util.logTag, "Got contents of %s: %s", diskFilePath, string(content))

		contents = append(contents, content)
	}

	err = util.unmount(tempDir)
	if err != nil {
		return [][]byte{}, err
	}

	return contents, nil
}

func (util diskUtil) unmount(tempDir string) error {
	util.logger.Debug(util.logTag, "Unmounting disk path '%s'", tempDir)

	_, err := util.mounter.Unmount(tempDir)
	if err != nil {
		return bosherr.WrapErrorf(err, "Unmounting '%s'", tempDir)
	}

	return nil
}
