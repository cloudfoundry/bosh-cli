package fakes

import (
	"fmt"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type JobInstallInput struct {
	Job bmrel.Job
}

type jobInstallOutput struct {
	err error
}

type FakeJobInstaller struct {
	JobInstallInputs []JobInstallInput
	installBehavior  map[string]jobInstallOutput
}

func NewFakeJobInstaller() *FakeJobInstaller {
	return &FakeJobInstaller{
		JobInstallInputs: []JobInstallInput{},
		installBehavior:  map[string]jobInstallOutput{},
	}
}

func (f *FakeJobInstaller) Install(job bmrel.Job) error {
	input := JobInstallInput{Job: job}
	f.JobInstallInputs = append(f.JobInstallInputs, input)
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	output, found := f.installBehavior[value]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: %s\nAvailible Behaviors: %s", value, f.installBehavior)
}

func (f *FakeJobInstaller) SetInstallBehavior(pkg bmrel.Job, targetDir string, err error) error {
	input := JobInstallInput{Job: pkg}
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.installBehavior[value] = jobInstallOutput{err: err}
	return nil
}
