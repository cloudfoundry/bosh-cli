package cmd

import (
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type AddBlobCmd struct {
	blobsDir boshreldir.BlobsDir
	fs       boshsys.FileSystem
	ui       boshui.UI
}

func NewAddBlobCmd(blobsDir boshreldir.BlobsDir, fs boshsys.FileSystem, ui boshui.UI) AddBlobCmd {
	return AddBlobCmd{blobsDir: blobsDir, fs: fs, ui: ui}
}

func (c AddBlobCmd) Run(opts AddBlobOpts) error {
	file, err := c.fs.OpenFile(opts.Args.Path, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening blob")
	}

	defer file.Close() //nolint:errcheck

	blob, err := c.blobsDir.TrackBlob(opts.Args.BlobsPath, file)
	if err != nil {
		return bosherr.WrapErrorf(err, "Tracking blob")
	}

	c.ui.PrintLinef("Added blob '%s'", blob.Path)

	return nil
}
