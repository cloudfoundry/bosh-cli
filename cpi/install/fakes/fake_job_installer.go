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

type jobInstallOutput struct {
	installedJob bmcpiinstall.InstalledJob
	err          error
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

func (f *FakeJobInstaller) Install(job bmrel.Job) (bmcpiinstall.InstalledJob, error) {
	input := JobInstallInput{Job: job}
	f.JobInstallInputs = append(f.JobInstallInputs, input)
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bmcpiinstall.InstalledJob{}, fmt.Errorf("Could not serialize input %#v", input)
	}
	output, found := f.installBehavior[value]

	if found {
		return output.installedJob, output.err
	}
	return bmcpiinstall.InstalledJob{}, fmt.Errorf("Unsupported Input: %s\nAvailible Behaviors: %s", value, f.installBehavior)
}

func (f *FakeJobInstaller) SetInstallBehavior(job bmrel.Job, installedJob bmcpiinstall.InstalledJob, err error) error {
	input := JobInstallInput{
		Job: job,
	}
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.installBehavior[value] = jobInstallOutput{installedJob: installedJob, err: err}
	return nil
}
