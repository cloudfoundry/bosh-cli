package fakes

import (
	"fmt"

	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type JobInstallInput struct {
	Job bmrel.Job
}

type jobInstallCallback func(job bmrel.Job) (bmcpiinstall.InstalledJob, error)

type FakeJobInstaller struct {
	JobInstallInputs []JobInstallInput
	installBehavior  map[string]jobInstallCallback
}

func NewFakeJobInstaller() *FakeJobInstaller {
	return &FakeJobInstaller{
		JobInstallInputs: []JobInstallInput{},
		installBehavior:  map[string]jobInstallCallback{},
	}
}

func (f *FakeJobInstaller) Install(job bmrel.Job) (bmcpiinstall.InstalledJob, error) {
	input := JobInstallInput{Job: job}
	f.JobInstallInputs = append(f.JobInstallInputs, input)
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bmcpiinstall.InstalledJob{}, fmt.Errorf("Could not serialize input %#v", input)
	}
	callback, found := f.installBehavior[value]

	if found {
		return callback(job)
	}
	return bmcpiinstall.InstalledJob{}, fmt.Errorf("Unsupported Input: %s\nAvailible Behaviors: %#v", value, f.installBehavior)
}

func (f *FakeJobInstaller) SetInstallBehavior(job bmrel.Job, callback jobInstallCallback) error {
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
