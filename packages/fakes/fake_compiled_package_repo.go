package fakes

import (
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/packages"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeCompiledPackageRepo struct {
	SavePackage bmrel.Package
	SaveRecord  bmpkgs.CompiledPackageRecord
	SaveError   error

	FindCompiledPackageRecord bmpkgs.CompiledPackageRecord
	FindCompiledPackageError  error
}

func NewFakeCompiledPackageRepo() *FakeCompiledPackageRepo {
	return &FakeCompiledPackageRepo{}
}

func (cpr *FakeCompiledPackageRepo) Save(
	pkg bmrel.Package,
	record bmpkgs.CompiledPackageRecord,
) error {
	cpr.SavePackage = pkg
	cpr.SaveRecord = record
	return cpr.SaveError
}

func (cpr *FakeCompiledPackageRepo) Find(pkg bmrel.Package) (bmpkgs.CompiledPackageRecord, bool, error) {
	return cpr.FindCompiledPackageRecord, cpr.FindCompiledPackageRecord != bmpkgs.CompiledPackageRecord{}, cpr.FindCompiledPackageError

}
