package cmd

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type CACertArg struct {
	Content string
	FS      boshsys.FileSystem
}

func (a *CACertArg) UnmarshalFlag(data string) error {
	if a.FS == nil {
		a.FS = boshsys.NewOsFileSystemWithStrictTempRoot(boshlog.NewLogger(boshlog.LevelNone))
	}

	if len(data) == 0 {
		return bosherr.Errorf("Expected CA cert to be non-empty")
	}

	if strings.Contains(data, "BEGIN") {
		a.Content = data
		return nil
	}

	absPath, err := a.FS.ExpandPath(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting absolute path '%s'", data)
	}

	content, err := a.FS.ReadFileString(absPath)
	if err != nil {
		return err
	}

	a.Content = content

	return nil
}
