package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeRelease struct {
	ReleaseName     string
	ReleaseVersion  string
	ReleaseJobs     []bmrel.Job
	ReleasePackages []*bmrel.Package

	FindJobByNameName  string
	FindJobByNameJob   bmrel.Job
	FindJobByNameFound bool

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

func (r *FakeRelease) Jobs() []bmrel.Job { return r.ReleaseJobs }

func (r *FakeRelease) Packages() []*bmrel.Package { return r.ReleasePackages }

func (r *FakeRelease) FindJobByName(jobName string) (bmrel.Job, bool) {
	for _, job := range r.ReleaseJobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return bmrel.Job{}, false
}

func (r *FakeRelease) Delete() error {
	r.DeleteCalled = true
	return r.DeleteErr
}

func (r *FakeRelease) Exists() bool {
	return !r.DeleteCalled
}
