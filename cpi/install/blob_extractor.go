package install

import (
	"os"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type BlobExtractor interface {
	Extract(blobID string, blobSHA1 string, targetDir string) error
}

type blobExtractor struct {
	fs         boshsys.FileSystem
	compressor boshcmd.Compressor
	blobstore  boshblob.Blobstore
	logger     boshlog.Logger
	logTag     string
}

func NewBlobExtractor(
	fs boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	logger boshlog.Logger,
) blobExtractor {
	return blobExtractor{
		fs:         fs,
		compressor: compressor,
		blobstore:  blobstore,
		logger:     logger,
		logTag:     "blobExtractor",
	}
}

func (e blobExtractor) Extract(blobID string, blobSHA1 string, targetDir string) error {
	filePath, err := e.blobstore.Get(blobID, blobSHA1)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting object from blobstore: %s", blobID)
	}
	defer e.cleanUpBlob(filePath)

	existed := e.fs.FileExists(targetDir)
	if !existed {
		err = e.fs.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating target dir: %s", targetDir)
		}
	}

	err = e.compressor.DecompressFileToDir(filePath, targetDir, boshcmd.CompressorOptions{})
	if err != nil {
		if !existed {
			e.cleanUpFile(targetDir)
		}
		return bosherr.WrapErrorf(err, "Extracting compiled package: BlobID:`%s', BlobSHA1: `%s'", blobID, blobSHA1)
	}
	return nil
}

func (e blobExtractor) cleanUpBlob(filePath string) {
	err := e.blobstore.CleanUp(filePath)
	if err != nil {
		e.logger.Error(
			e.logTag,
			bosherr.WrapErrorf(err, "Removing compiled package tarball: %s", filePath).Error(),
		)
	}
}

func (e blobExtractor) cleanUpFile(filePath string) {
	err := e.fs.RemoveAll(filePath)
	if err != nil {
		e.logger.Error(
			e.logTag,
			bosherr.WrapErrorf(err, "Removing: %s", filePath).Error(),
		)
	}
}
