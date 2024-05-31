package releasedir

import (
	"io"

	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshrelman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . ReleaseDir

type ReleaseDir interface {
	Init(bool) error
	Reset() error

	GenerateJob(string) error
	GeneratePackage(string) error

	// DefaultName returns a string for the release.
	DefaultName() (string, error)

	// NextDevVersion and NextFinalVersion returns a next version for the that name.
	// It does not account for gaps and just plainly increments.
	NextDevVersion(name string, timestamp bool) (semver.Version, error)
	NextFinalVersion(name string) (semver.Version, error)

	// FindRelease returns last dev or final release version if it's empty;
	// otherwise it finds a release by given name and version.
	FindRelease(name string, version semver.Version) (boshrel.Release, error)

	// BuildRelease builds a new version of the Release
	// from the release directory by looking at jobs, packages, etc. directories.
	BuildRelease(name string, version semver.Version, force bool) (boshrel.Release, error)
	VendorPackage(pkg *boshpkg.Package, prefix string) error

	// FinalizeRelease adds the Release to the final list so that it's consumable by others.
	FinalizeRelease(release boshrel.Release, force bool) error
}

//counterfeiter:generate . Config

type Config interface {
	Name() (string, error)
	SaveName(string) error

	Blobstore() (string, map[string]interface{}, error)
}

//counterfeiter:generate . Generator

type Generator interface {
	GenerateJob(string) error
	GeneratePackage(string) error
}

//counterfeiter:generate . GitRepo

type GitRepo interface {
	Init() error
	LastCommitSHA() (string, error)
	MustNotBeDirty(force bool) (dirty bool, err error)
}

//counterfeiter:generate . BlobsDir

type BlobsDir interface {
	Init() error
	Blobs() ([]Blob, error)

	SyncBlobs(numOfParallelWorkers int) error
	UploadBlobs() error

	TrackBlob(string, io.ReadCloser) (Blob, error)
	UntrackBlob(string) error
}

//counterfeiter:generate . BlobsDirReporter

type BlobsDirReporter interface {
	BlobDownloadStarted(path string, size int64, blobID, sha1 string)
	BlobDownloadFinished(path, blobID string, err error)

	BlobUploadStarted(path string, size int64, sha1 string)
	BlobUploadFinished(path, blobID string, err error)
}

type Blob struct {
	Path string
	Size int64

	BlobstoreID string
	SHA1        string
}

//counterfeiter:generate . ReleaseIndex

type ReleaseIndex interface {
	LastVersion(name string) (*semver.Version, error)

	Contains(boshrel.Release) (bool, error)
	Add(boshrelman.Manifest) error

	ManifestPath(name, version string) string
}

//counterfeiter:generate . ReleaseIndexReporter

type ReleaseIndexReporter interface {
	ReleaseIndexAdded(name, desc string, err error)
}

//counterfeiter:generate . DigestBlobstore

type DigestBlobstore interface {
	Get(blobID string, digest boshcrypto.Digest) (fileName string, err error)
	CleanUp(fileName string) (err error)
	Create(fileName string) (blobID string, digest boshcrypto.MultipleDigest, err error)
	Validate() (err error)
	Delete(blobId string) (err error)
}
