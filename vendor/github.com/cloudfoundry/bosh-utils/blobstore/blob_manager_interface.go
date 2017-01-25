package blobstore

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	"io"
)

type BlobManagerInterface interface {
	Fetch(blobID string) (boshsys.File, error, int)

	Write(blobID string, reader io.Reader) error

	GetPath(blobID string, digest boshcrypto.Digest) (string, error)

	Delete(blobID string) error

	BlobExists(blobID string) bool
}
