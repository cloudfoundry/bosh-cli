package fakes

import (
	"fmt"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type JobInstallInput struct {
	Job   bmreljob.Job
	Stage bmeventlog.Stage
}

type jobInstallCallback func(job bmreljob.Job, stage bmeventlog.Stage) (bminstalljob.InstalledJob, error)

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

func (f *FakeInstaller) Install(job bmreljob.Job, stage bmeventlog.Stage) (bminstalljob.InstalledJob, error) {
	input := JobInstallInput{
		Job:   job,
		Stage: stage,
	}
	f.JobInstallInputs = append(f.JobInstallInputs, input)
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bminstalljob.InstalledJob{}, fmt.Errorf("Could not serialize input %#v", input)
	}
	callback, found := f.installBehavior[value]

	if found {
		return callback(job, stage)
	}
	return bminstalljob.InstalledJob{}, fmt.Errorf("Unsupported Input: %s\nAvailible Behaviors: %#v", value, f.installBehavior)
}

func (f *FakeInstaller) SetInstallBehavior(job bmreljob.Job, stage bmeventlog.Stage, callback jobInstallCallback) error {
	input := JobInstallInput{
		Job:   job,
		Stage: stage,
	}
	value, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return fmt.Errorf("Could not serialize input %#v", input)
	}
	f.installBehavior[value] = callback
	return nil
}
