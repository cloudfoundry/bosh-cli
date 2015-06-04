package blob

import (
	"os"

	boshblob "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
)

type Extractor interface {
	Extract(blobID string, blobSHA1 string, targetDir string) error
}

type extractor struct {
	fs         boshsys.FileSystem
	compressor boshcmd.Compressor
	blobstore  boshblob.Blobstore
	logger     boshlog.Logger
	logTag     string
}

func NewExtractor(
	fs boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	logger boshlog.Logger,
) Extractor {
	return extractor{
		fs:         fs,
		compressor: compressor,
		blobstore:  blobstore,
		logger:     logger,
		logTag:     "blobExtractor",
	}
}

func (e extractor) Extract(blobID string, blobSHA1 string, targetDir string) error {
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
		return bosherr.WrapErrorf(err, "Extracting compiled package: BlobID:'%s', BlobSHA1: '%s'", blobID, blobSHA1)
	}
	return nil
}

func (e extractor) cleanUpBlob(filePath string) {
	err := e.blobstore.CleanUp(filePath)
	if err != nil {
		e.logger.Error(
			e.logTag,
			bosherr.WrapErrorf(err, "Removing compiled package tarball: %s", filePath).Error(),
		)
	}
}

func (e extractor) cleanUpFile(filePath string) {
	err := e.fs.RemoveAll(filePath)
	if err != nil {
		e.logger.Error(
			e.logTag,
			bosherr.WrapErrorf(err, "Removing: %s", filePath).Error(),
		)
	}
}
