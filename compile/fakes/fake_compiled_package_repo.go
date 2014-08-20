package fakes

import (
	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeCompiledPackageRepo struct {
	SavePackage bmrel.Package
	SaveRecord  bmcomp.CompiledPackageRecord
	SaveError   error

	FindCompiledPackageRecord bmcomp.CompiledPackageRecord
	FindCompiledPackageError  error
}

func NewFakeCompiledPackageRepo() *FakeCompiledPackageRepo {
	return &FakeCompiledPackageRepo{}
}

func (cpr *FakeCompiledPackageRepo) Save(
	pkg bmrel.Package,
	record bmcomp.CompiledPackageRecord,
) error {
	cpr.SavePackage = pkg
	cpr.SaveRecord = record
	return cpr.SaveError
}

func (cpr *FakeCompiledPackageRepo) Find(pkg bmrel.Package) (bmcomp.CompiledPackageRecord, bool, error) {
	return cpr.FindCompiledPackageRecord, cpr.FindCompiledPackageRecord != bmcomp.CompiledPackageRecord{}, cpr.FindCompiledPackageError

}
