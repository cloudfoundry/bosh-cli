package templatescompiler

import (
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/property"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
)

type JobListRenderer interface {
	Render(
		releaseJobs []bireljob.Job,
		jobProperties biproperty.Map,
		globalProperties biproperty.Map,
		deploymentName string,
	) (RenderedJobList, error)
}

type jobListRenderer struct {
	jobRenderer JobRenderer
	logger      boshlog.Logger
	logTag      string
}

func NewJobListRenderer(
	jobRenderer JobRenderer,
	logger boshlog.Logger,
) JobListRenderer {
	return &jobListRenderer{
		jobRenderer: jobRenderer,
		logger:      logger,
		logTag:      "jobListRenderer",
	}
}

func (r *jobListRenderer) Render(
	releaseJobs []bireljob.Job,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
) (RenderedJobList, error) {
	r.logger.Debug(r.logTag, "Rendering job list: deploymentName='%s' jobProperties=%#v globalProperties=%#v", deploymentName, jobProperties, globalProperties)
	renderedJobList := NewRenderedJobList()

	// render all the jobs' templates
	for _, releaseJob := range releaseJobs {
		renderedJob, err := r.jobRenderer.Render(releaseJob, jobProperties, globalProperties, deploymentName)
		if err != nil {
			defer renderedJobList.DeleteSilently()
			return renderedJobList, bosherr.WrapErrorf(err, "Rendering templates for job '%s/%s'", releaseJob.Name, releaseJob.Fingerprint)
		}
		renderedJobList.Add(renderedJob)
	}

	return renderedJobList, nil
}
