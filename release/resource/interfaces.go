package resource

import (
	crypto2 "github.com/cloudfoundry/bosh-utils/crypto"

	"github.com/cloudfoundry/bosh-cli/v7/crypto"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Archive

type Archive interface {
	Fingerprint() (string, error)
	Build(expectedFp string) (string, string, error)
	CleanUp(path string)
}

type ArchiveFunc func(args ArchiveFactoryArgs) Archive

type ArchiveFactoryArgs struct {
	Files          []File
	PrepFiles      []File
	Chunks         []string
	FollowSymlinks bool
}

//counterfeiter:generate . ArchiveIndex

type ArchiveIndex interface {
	Find(name, fingerprint string) (string, string, error)
	Add(name, fingerprint, path, sha1 string) (string, string, error)
}

//counterfeiter:generate . Resource

type Resource interface {
	Name() string
	Prefix(prefix string)
	Fingerprint() string

	ArchivePath() string
	ArchiveDigest() string

	Build(dev, final ArchiveIndex) error
	Finalize(final ArchiveIndex) error

	RehashWithCalculator(calculator crypto.DigestCalculator, archiveFilePathReader crypto2.ArchiveDigestFilePathReader) (Resource, error)
}

//counterfeiter:generate . Fingerprinter

type Fingerprinter interface {
	Calculate([]File, []string) (string, error)
}
