package crypto

import (
	"io"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Digest interface {
	Verify(io.Reader) error
	VerifyFilePath(filePath string, fs boshsys.FileSystem) error
	Algorithm() Algorithm
	String() string
}

var _ Digest = digestImpl{}

type Algorithm interface {
	CreateDigest(io.Reader) (Digest, error)
	Name() string
}

var _ Algorithm = algorithmSHAImpl{}
var _ Algorithm = unknownAlgorithmImpl{}
