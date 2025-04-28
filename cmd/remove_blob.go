package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type RemoveBlobCmd struct {
	blobsDir boshreldir.BlobsDir
	ui       biui.UI
}

func NewRemoveBlobCmd(blobsDir boshreldir.BlobsDir, ui biui.UI) RemoveBlobCmd {
	return RemoveBlobCmd{blobsDir: blobsDir, ui: ui}
}

func (c RemoveBlobCmd) Run(opts RemoveBlobOpts) error {
	err := c.blobsDir.UntrackBlob(opts.Args.BlobsPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Untracking blob")
	}

	c.ui.PrintLinef("Removed blob '%s'", opts.Args.BlobsPath)

	return nil
}
