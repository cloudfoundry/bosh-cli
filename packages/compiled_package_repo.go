package packages

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	"fmt"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	"sort"
	"strings"
)

type CompiledPackageRecord struct {
	BlobID   string
	BlobSha1 string
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
	PackageName string
	// Fingerprint of a package captures the sorted names of its dependencies
	// (but not the dependencies' fingerprints)
	PackageFingerprint string
	DependencyKey      string
}

func (cpr compiledPackageRepo) pkgKey(pkg bmrel.Package) packageToCompiledPackageKey {
	return packageToCompiledPackageKey{
		PackageName:        pkg.Name,
		PackageFingerprint: pkg.Fingerprint,
		DependencyKey:      cpr.convertToDependencyKey(ResolveDependencies(&pkg)),
	}
}

func (cpr compiledPackageRepo) convertToDependencyKey(packages []*bmrel.Package) string {
	dependencyKeys := []string{}
	for _, pkg := range packages {
		dependencyKeys = append(dependencyKeys, fmt.Sprintf("%s:%s", pkg.Name, pkg.Fingerprint))
	}
	sort.Strings(dependencyKeys)
	return strings.Join(dependencyKeys, ",")
}
