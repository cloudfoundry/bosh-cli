package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FileArg struct {
	ExpandedPath string
	FS           boshsys.FileSystem
}

func (a *FileArg) UnmarshalFlag(data string) error {
	expandedPath, err := a.FS.ExpandPath(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "checking file path")
	}
	a.ExpandedPath = expandedPath

	if a.FS.FileExists(expandedPath) {
		stat, err := a.FS.Stat(expandedPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "checking file path")
		}

		if stat.IsDir() {
			return bosherr.Errorf("path must not be directory")
		}
	}

	return nil
}
