package blobstore

import (
	"io"
	"os"

	boshdavcli "github.com/cloudfoundry/bosh-agent/davcli/client"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

type Blobstore interface {
	Get(blobID, destinationPath string) error
	Add(sourcePath string) (blobID string, err error)
}

type Config struct {
	Endpoint string
	Username string
	Password string
}

type blobstore struct {
	davClient     boshdavcli.Client
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
	logger        boshlog.Logger
	logTag        string
}

func NewBlobstore(davClient boshdavcli.Client, uuidGenerator boshuuid.Generator, fs boshsys.FileSystem, logger boshlog.Logger) Blobstore {
	return &blobstore{
		davClient:     davClient,
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
		logTag:        "blobstore",
	}
}

func (b *blobstore) Get(blobID, destinationPath string) error {
	b.logger.Debug(b.logTag, "Downloading blob %s to %s", blobID, destinationPath)

	readCloser, err := b.davClient.Get(blobID)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting blob %s from blobstore", blobID)
	}
	defer readCloser.Close()

	targetFile, err := b.fs.OpenFile(destinationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening file for blob at %s", destinationPath)
	}

	_, err = io.Copy(targetFile, readCloser)
	if err != nil {
		return bosherr.WrapErrorf(err, "Saving blob to %s", destinationPath)
	}

	return nil
}

func (b *blobstore) Add(sourcePath string) (blobID string, err error) {
	blobID, err = b.uuidGenerator.Generate()
	if err != nil {
		return "", bosherr.WrapError(err, "Generating Blob ID")
	}

	b.logger.Debug(b.logTag, "Uploading blob %s from %s", blobID, sourcePath)

	file, err := b.fs.OpenFile(sourcePath, os.O_RDONLY, 0)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Opening file for reading %s", sourcePath)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Getting fileInfo from %s", sourcePath)
	}

	err = b.davClient.Put(blobID, file, fileInfo.Size())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Putting file '%s' into blobstore (via DAVClient) as blobID '%s'", sourcePath, blobID)
	}

	return blobID, nil
}
