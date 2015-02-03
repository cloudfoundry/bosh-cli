package templatescompiler

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
)

type RenderedJob interface {
	Job() bmreljob.Job
	Path() string // dir of multiple rendered files
	Delete() error
	DeleteSilently()
}

type renderedJob struct {
	job    bmreljob.Job
	path   string
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

func NewRenderedJob(
	job bmreljob.Job,
	path string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) RenderedJob {
	return &renderedJob{
		job:    job,
		path:   path,
		fs:     fs,
		logger: logger,
		logTag: "renderedJob",
	}
}

func (j *renderedJob) Job() bmreljob.Job { return j.job }

// Path returns a parent directory with one or more sub-dirs for each job, each with one or more rendered template files
func (j *renderedJob) Path() string { return j.path }

//TODO: test me
func (j *renderedJob) Delete() error {
	err := j.fs.RemoveAll(j.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting rendered job '%s' tarball '%s'", j.job.Name, j.path)
	}
	return nil
}

//TODO: test me
func (j *renderedJob) DeleteSilently() {
	err := j.Delete()
	if err != nil {
		j.logger.Error(j.logTag, "Failed to delete rendered job: %s", err.Error())
	}
}

func (j *renderedJob) String() string {
	return fmt.Sprintf("renderedJob{job: '%s', path: '%s'}", j.job.Name, j.path)
}
