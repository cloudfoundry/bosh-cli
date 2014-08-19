package blobstore

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshdir "github.com/cloudfoundry/bosh-agent/settings/directories"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	"path/filepath"
)

type BlobManager struct {
	fs          boshsys.FileSystem
	dirProvider boshdir.Provider
}

func NewBlobManager(fs boshsys.FileSystem, dirProvider boshdir.Provider) (manager BlobManager) {
	manager.fs = fs
	manager.dirProvider = dirProvider
	return
}

func (manager BlobManager) Fetch(blobID string) (blobBytes []byte, err error) {
	blobPath := filepath.Join(manager.dirProvider.MicroStore(), blobID)

	blobBytes, err = manager.fs.ReadFile(blobPath)
	if err != nil {
		err = bosherr.WrapError(err, "Reading blob")
	}
	return
}

func (manager BlobManager) Write(blobID string, blobBytes []byte) (err error) {
	blobPath := filepath.Join(manager.dirProvider.MicroStore(), blobID)

	err = manager.fs.WriteFile(blobPath, blobBytes)
	if err != nil {
		err = bosherr.WrapError(err, "Updating blob")
	}
	return
}
