package release

import (
	boshjob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	boshlic "github.com/cloudfoundry/bosh-cli/v7/release/license"
	boshman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	boshres "github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

// You only need **one** of these per package!
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type Extractor interface {
	Extract(string) (Release, error)
}

//counterfeiter:generate . Reader

type Reader interface {
	// Read reads an archive for example and returns a Release.
	Read(string) (Release, error)
}

//counterfeiter:generate . Writer

type Writer interface {
	// Write writes an archive for example and returns its path.
	// Archive does not include packages that have fingerprints
	// included in the second argument.
	Write(Release, []string) (string, error)
}

//counterfeiter:generate . Release

type Release interface {
	Name() string
	SetName(string)

	Version() string
	SetVersion(string)

	CommitHashWithMark(string) string
	SetCommitHash(string)
	SetUncommittedChanges(bool)

	Jobs() []*boshjob.Job
	Packages() []*boshpkg.Package
	CompiledPackages() []*boshpkg.CompiledPackage
	License() *boshlic.License

	IsCompiled() bool

	FindJobByName(string) (boshjob.Job, bool)
	Manifest() boshman.Manifest

	Build(dev, final ArchiveIndicies, parallel int) error
	Finalize(final ArchiveIndicies, parallel int) error

	CopyWith(jobs []*boshjob.Job,
		packages []*boshpkg.Package,
		lic *boshlic.License,
		compiledPackages []*boshpkg.CompiledPackage) Release

	CleanUp() error
}

type ArchiveIndicies struct {
	Jobs     boshres.ArchiveIndex
	Packages boshres.ArchiveIndex
	Licenses boshres.ArchiveIndex
}

type Manager interface {
	Add(Release)
	List() []Release
	Find(string) (Release, bool)
	DeleteAll() error
}
