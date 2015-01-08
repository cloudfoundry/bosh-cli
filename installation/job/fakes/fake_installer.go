package fakes

import (
	"fmt"

	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type JobInstallInput struct {
	Job bmrel.Job
}

type jobInstallCallback func(job bmrel.Job) (bminstalljob.InstalledJob, error)

type FakeInstaller struct {
	JobInstallInputs []JobInstallInput
	installBehavior  map[string]jobInstallCallback
}

func NewFakeInstaller() *FakeInstaller {
	return &FakeInstaller{
		JobInstallInputs: []JobInstallInput{},
		installBehavior:  map[string]jobInstallCallback{},
	}
}

func (f *FakeInstaller) Install(job bmrel.Job) (bminstalljob.InstalledJob, error) {
	input := JobInstallInput{Job: job}
	f.JobInstallInputs = append(f.JobInstallInputs, input)
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bminstalljob.InstalledJob{}, fmt.Errorf("Could not serialize input %#v", input)
	}
	callback, found := f.installBehavior[value]

	if found {
		return callback(job)
	}
	return bminstalljob.InstalledJob{}, fmt.Errorf("Unsupported Input: %s\nAvailible Behaviors: %#v", value, f.installBehavior)
}

func (f *FakeInstaller) SetInstallBehavior(job bmrel.Job, callback jobInstallCallback) error {
	input := JobInstallInput{
		Job: job,
	}
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.installBehavior[value] = callback
	return nil
}
