package release

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type release struct {
	name          string
	version       string
	jobs          []bmreljob.Job
	packages      []*bmrelpkg.Package
	extractedPath string
	fs            boshsys.FileSystem
}

type Release interface {
	Name() string
	Version() string
	Jobs() []bmreljob.Job
	Packages() []*bmrelpkg.Package
	FindJobByName(jobName string) (job bmreljob.Job, found bool)
	Delete() error
	Exists() bool
}

func NewRelease(
	name string,
	version string,
	jobs []bmreljob.Job,
	packages []*bmrelpkg.Package,
	extractedPath string,
	fs boshsys.FileSystem,
) Release {
	return &release{
		name:          name,
		version:       version,
		jobs:          jobs,
		packages:      packages,
		extractedPath: extractedPath,
		fs:            fs,
	}
}

func (r *release) Name() string { return r.name }

func (r *release) Version() string { return r.version }

func (r *release) Jobs() []bmreljob.Job { return r.jobs }

func (r *release) Packages() []*bmrelpkg.Package { return r.packages }

func (r *release) FindJobByName(jobName string) (bmreljob.Job, bool) {
	for _, job := range r.jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return bmreljob.Job{}, false
}

// Delete removes the extracted release code.
// Since packages and jobs are under the same path, they will be deleted too.
func (r *release) Delete() error {
	return r.fs.RemoveAll(r.extractedPath)
}

// Exists returns false after Delete (or if extractedPath does not exist)
func (r *release) Exists() bool {
	return r.fs.FileExists(r.extractedPath)
}
