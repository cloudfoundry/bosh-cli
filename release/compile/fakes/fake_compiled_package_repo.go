package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelcomp "github.com/cloudfoundry/bosh-micro-cli/release/compile"
)

type FakeCompiledPackageRepo struct {
	SavePackage bmrel.Package
	SaveRecord  bmrelcomp.CompiledPackageRecord
	SaveError   error
}

func NewFakeCompiledPackageRepo() *FakeCompiledPackageRepo {
	return &FakeCompiledPackageRepo{}
}

func (cpr *FakeCompiledPackageRepo) Save(
	pkg bmrel.Package,
	record bmrelcomp.CompiledPackageRecord,
) error {
	cpr.SavePackage = pkg
	cpr.SaveRecord = record
	return cpr.SaveError
}

func (cpr *FakeCompiledPackageRepo) Find(pkg bmrel.Package) (bmrelcomp.CompiledPackageRecord, bool, error) {
	return bmrelcomp.CompiledPackageRecord{}, false, nil
}
