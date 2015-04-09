package fakes

import (
	bmreljob "github.com/cloudfoundry/bosh-init/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
)

type FakeRelease struct {
	ReleaseName     string
	ReleaseVersion  string
	ReleaseJobs     []bmreljob.Job
	ReleasePackages []*bmrelpkg.Package

	DeleteCalled bool
	DeleteErr    error
}

func NewFakeRelease() *FakeRelease {
	return &FakeRelease{}
}

func New(name, version string) *FakeRelease {
	return &FakeRelease{
		ReleaseName:    name,
		ReleaseVersion: version,
	}
}

func (r *FakeRelease) Name() string { return r.ReleaseName }

func (r *FakeRelease) Version() string { return r.ReleaseVersion }

func (r *FakeRelease) Jobs() []bmreljob.Job { return r.ReleaseJobs }

func (r *FakeRelease) Packages() []*bmrelpkg.Package { return r.ReleasePackages }

func (r *FakeRelease) FindJobByName(jobName string) (bmreljob.Job, bool) {
	for _, job := range r.ReleaseJobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return bmreljob.Job{}, false
}

func (r *FakeRelease) Delete() error {
	r.DeleteCalled = true
	return r.DeleteErr
}

func (r *FakeRelease) Exists() bool {
	return !r.DeleteCalled
}
