package crypto

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"io"
	"os"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Digest interface {
	Verify(io.Reader) error
	VerifyFilePath(filePath string, fs boshsys.FileSystem) error
	Algorithm() Algorithm
	String() string
}

//counterfeiter:generate . ArchiveDigestFilePathReader
type ArchiveDigestFilePathReader interface {
	OpenFile(path string, flag int, perm os.FileMode) (boshsys.File, error)
}

var _ Digest = digestImpl{}
var _ Digest = MultipleDigest{}

type Algorithm interface {
	CreateDigest(io.Reader) (Digest, error)
	Name() string
}

var _ Algorithm = algorithmSHAImpl{}
var _ Algorithm = unknownAlgorithmImpl{}
