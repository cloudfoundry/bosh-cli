package release

import (
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type Release struct {
	Name    string
	Version string

	CommitHash         string
	UncommittedChanges bool

	Jobs          []bmreljob.Job
	Packages      []*Package
	ExtractedPath string
	TarballPath   string
}

type Package struct {
	Name          string
	Version       string
	Fingerprint   string
	Sha1          string
	Dependencies  []*Package
	ExtractedPath string
}

func (r Release) FindJobByName(jobName string) (bmreljob.Job, bool) {
	for _, job := range r.Jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return bmreljob.Job{}, false
}
