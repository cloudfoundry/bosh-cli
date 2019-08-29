package opts

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FileBytesWithPathArg struct {
	FS boshsys.FileSystem

	Bytes []byte
	Path  string
}

func (a *FileBytesWithPathArg) UnmarshalFlag(data string) error {
	if a.FS == nil {
		a.FS = boshsys.NewOsFileSystemWithStrictTempRoot(boshlog.NewLogger(boshlog.LevelNone))
	}

	if len(data) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	absPath, err := a.FS.ExpandPath(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting absolute path '%s'", data)
	}

	bytes, err := a.FS.ReadFile(absPath)
	if err != nil {
		return err
	}

	(*a).Bytes = bytes
	(*a).Path = absPath

	return nil
}
