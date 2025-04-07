package releasedir

import (
	gobytes "bytes"
	"encoding/json"
	"os"

	s3client "github.com/cloudfoundry/bosh-s3cli/client"
	s3config "github.com/cloudfoundry/bosh-s3cli/config"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type S3Blobstore struct {
	fs      boshsys.FileSystem
	uuidGen boshuuid.Generator
	options map[string]interface{}
}

func NewS3Blobstore(
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	options map[string]interface{},
) S3Blobstore {
	return S3Blobstore{
		fs:      fs,
		uuidGen: uuidGen,
		options: options,
	}
}

func (b S3Blobstore) Get(blobID string) (string, error) {
	client, err := b.client()
	if err != nil {
		return "", err
	}

	file, err := b.fs.TempFile("bosh-s3-blob")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating destination file")
	}

	defer file.Close() //nolint:errcheck

	err = client.Get(blobID, file)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (b S3Blobstore) Create(path string) (string, error) {
	client, err := b.client()
	if err != nil {
		return "", err
	}

	blobID, err := b.uuidGen.Generate()
	if err != nil {
		return "", bosherr.WrapError(err, "Generating blobstore ID")
	}

	file, err := b.fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", bosherr.WrapError(err, "Opening source file")
	}

	defer file.Close() //nolint:errcheck

	err = client.Put(file, blobID)
	if err != nil {
		return "", bosherr.WrapError(err, "Generating blobstore ID")
	}

	return blobID, nil
}

func (b S3Blobstore) CleanUp(path string) error {
	return b.fs.RemoveAll(path)
}

func (b S3Blobstore) Delete(blobID string) error {
	panic("Not implemented")
}

func (b S3Blobstore) Validate() error {
	_, err := b.client()
	return err
}

func (b S3Blobstore) client() (s3client.S3CompatibleClient, error) {
	bytes, err := json.Marshal(b.options)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Marshaling config")
	}

	conf, err := s3config.NewFromReader(gobytes.NewBuffer(bytes))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading config")
	}

	s3ClientSDK, err := s3client.NewAwsS3Client(&conf)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Building client SDK")
	}

	client := s3client.New(s3ClientSDK, &conf)

	return client, nil
}
