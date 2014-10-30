package blobstore

import (
	"io"
	"os"

	boshdavcli "github.com/cloudfoundry/bosh-agent/davcli/client"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Blobstore interface {
	Get(string, string) error
	Save(string, string) error
}

type Config struct {
	Endpoint string
	Username string
	Password string
}

type blobstore struct {
	davClient boshdavcli.Client
	fs        boshsys.FileSystem
	logger    boshlog.Logger
	logTag    string
}

func NewBlobstore(davClient boshdavcli.Client, fs boshsys.FileSystem, logger boshlog.Logger) Blobstore {
	return &blobstore{
		davClient: davClient,
		fs:        fs,
		logger:    logger,
		logTag:    "blobstore",
	}
}

func (b *blobstore) Get(blobID string, destinationPath string) error {
	b.logger.Debug(b.logTag, "Downloading blob %s to %s", blobID, destinationPath)

	readCloser, err := b.davClient.Get(blobID)
	if err != nil {
		return bosherr.WrapError(err, "Getting blob %s from blobstore", blobID)
	}
	defer readCloser.Close()

	targetFile, err := b.fs.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return bosherr.WrapError(err, "Opening file for blob at %s", destinationPath)
	}

	_, err = io.Copy(targetFile, readCloser)
	if err != nil {
		return bosherr.WrapError(err, "Saving blob to %s", destinationPath)
	}

	return nil
}

func (b *blobstore) Save(sourcePath string, blobID string) error {
	b.logger.Debug(b.logTag, "Uploading blob %s from %s", blobID, sourcePath)

	file, err := b.fs.OpenFile(sourcePath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapError(err, "Opening file for reading %s", sourcePath)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return bosherr.WrapError(err, "Getting fileInfo from %s", sourcePath)
	}

	return b.davClient.Put(blobID, file, fileInfo.Size())
}
