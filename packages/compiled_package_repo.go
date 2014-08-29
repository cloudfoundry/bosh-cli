package packages

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type CompiledPackageRecord struct {
	BlobID      string `yaml:"blob_id"`
	Fingerprint string `yaml:"fingerprint"`
}

type CompiledPackageRepo interface {
	Save(bmrel.Package, CompiledPackageRecord) error
	Find(bmrel.Package) (CompiledPackageRecord, bool, error)
}

type compiledPackageRepo struct {
	index bmindex.Index
}

func NewCompiledPackageRepo(index bmindex.Index) CompiledPackageRepo {
	return &compiledPackageRepo{index: index}
}

func (cpr *compiledPackageRepo) Save(pkg bmrel.Package, record CompiledPackageRecord) error {
	err := cpr.index.Save(cpr.pkgKey(pkg), record)

	if err != nil {
		return bosherr.WrapError(err, "Saving compiled package")
	}

	return nil
}

func (cpr *compiledPackageRepo) Find(pkg bmrel.Package) (CompiledPackageRecord, bool, error) {
	var record CompiledPackageRecord

	err := cpr.index.Find(cpr.pkgKey(pkg), &record)
	if err != nil {
		if err == bmindex.ErrNotFound {
			return record, false, nil
		}

		return record, false, bosherr.WrapError(err, "Finding compiled package")
	}

	return record, true, nil
}

type packageToCompiledPackageKey struct {
	PackageName    string
	PackageVersion string

	// Fingerprint of a package captures its dependenices
	PackageFingerprint string
}

func (cpr compiledPackageRepo) pkgKey(pkg bmrel.Package) packageToCompiledPackageKey {
	return packageToCompiledPackageKey{
		PackageName:        pkg.Name,
		PackageVersion:     pkg.Version,
		PackageFingerprint: pkg.Fingerprint,
	}
}
