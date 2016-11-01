package blobstore

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"io"
)

type BlobManagerInterface interface {
	Fetch(blobID string) (boshsys.File, error, int)

	Write(blobID string, reader io.Reader) error

	GetPath(blobID string) (string, error)

	Delete(blobID string) error
}
