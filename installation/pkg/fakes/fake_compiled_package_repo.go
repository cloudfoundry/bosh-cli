package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type SaveInput struct {
	Package bmrel.Package
	Record  bmpkgs.CompiledPackageRecord
}
type saveOutput struct {
	err error
}
type FindInput struct {
	Package bmrel.Package
}
type findOutput struct {
	record bmpkgs.CompiledPackageRecord
	found  bool
	err    error
}

type FakeCompiledPackageRepo struct {
	SaveInputs []SaveInput
	FindInputs []FindInput

	saveBehavior map[string]saveOutput
	findBehavior map[string]findOutput
}

func NewFakeCompiledPackageRepo() *FakeCompiledPackageRepo {
	return &FakeCompiledPackageRepo{
		SaveInputs:   []SaveInput{},
		FindInputs:   []FindInput{},
		saveBehavior: map[string]saveOutput{},
		findBehavior: map[string]findOutput{},
	}
}

func (cpr *FakeCompiledPackageRepo) Save(pkg bmrel.Package, record bmpkgs.CompiledPackageRecord) error {
	input := SaveInput{Package: pkg, Record: record}
	cpr.SaveInputs = append(cpr.SaveInputs, input)

	inputString, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling Save input")
	}
	output, found := cpr.saveBehavior[inputString]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: Save('%#v', '%#v')", pkg, record)
}

func (cpr *FakeCompiledPackageRepo) SetSaveBehavior(pkg bmrel.Package, record bmpkgs.CompiledPackageRecord, err error) error {
	input := SaveInput{Package: pkg, Record: record}
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}
	cpr.saveBehavior[inputString] = saveOutput{err: err}
	return nil
}

func (cpr *FakeCompiledPackageRepo) Find(pkg bmrel.Package) (bmpkgs.CompiledPackageRecord, bool, error) {
	input := FindInput{Package: pkg}
	cpr.FindInputs = append(cpr.FindInputs, input)

	inputString, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bmpkgs.CompiledPackageRecord{}, false, bosherr.WrapError(err, "Marshaling Find input")
	}
	output, found := cpr.findBehavior[inputString]

	if found {
		return output.record, output.found, output.err
	}
	return bmpkgs.CompiledPackageRecord{}, false, fmt.Errorf("Unsupported input: Find('%#v')", pkg)
}

func (cpr *FakeCompiledPackageRepo) SetFindBehavior(pkg bmrel.Package, record bmpkgs.CompiledPackageRecord, found bool, err error) error {
	input := FindInput{Package: pkg}
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}
	cpr.findBehavior[inputString] = findOutput{record: record, found: found, err: err}
	return nil
}
