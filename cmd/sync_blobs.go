package cmd

import (
	boshreldir "github.com/cloudfoundry/bosh-init/releasedir"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type SyncBlobsCmd struct {
	blobsDir boshreldir.BlobsDir
}

func NewSyncBlobsCmd(blobsDir boshreldir.BlobsDir) SyncBlobsCmd {
	return SyncBlobsCmd{blobsDir: blobsDir}
}

func (c SyncBlobsCmd) Run() error {
	err := c.blobsDir.DownloadBlobs()
	if err != nil {
		return bosherr.WrapErrorf(err, "Downloading blobs")
	}

	return nil
}
