package blobstore

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"path/filepath"
)

type BlobManager struct {
	fs            boshsys.FileSystem
	blobstorePath string
}

func NewBlobManager(fs boshsys.FileSystem, blobstorePath string) (manager BlobManager) {
	manager.fs = fs
	manager.blobstorePath = blobstorePath
	return
}

func (manager BlobManager) Fetch(blobID string) (blobBytes []byte, err error) {
	blobPath := filepath.Join(manager.blobstorePath, blobID)

	blobBytes, err = manager.fs.ReadFile(blobPath)
	if err != nil {
		err = bosherr.WrapError(err, "Reading blob")
	}
	return
}

func (manager BlobManager) Write(blobID string, blobBytes []byte) (err error) {
	blobPath := filepath.Join(manager.blobstorePath, blobID)

	err = manager.fs.WriteFile(blobPath, blobBytes)
	if err != nil {
		err = bosherr.WrapError(err, "Updating blob")
	}
	return
}
