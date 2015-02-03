package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
)

type JobResolver interface {
	Resolve(jobName, releaseName string) (bmreljob.Job, error)
}

type resolver struct {
	releaseSetResolver bmrelset.Resolver
}

func NewJobResolver(releaseSetResolver bmrelset.Resolver) JobResolver {
	return &resolver{
		releaseSetResolver: releaseSetResolver,
	}
}

func (r *resolver) Resolve(jobName, releaseName string) (bmreljob.Job, error) {
	release, err := r.releaseSetResolver.Find(releaseName)
	if err != nil {
		return bmreljob.Job{}, bosherr.WrapErrorf(err, "Resolving release '%s'", releaseName)
	}

	releaseJob, found := release.FindJobByName(jobName)
	if !found {
		return bmreljob.Job{}, bosherr.Errorf("Finding job '%s' in release '%s'", jobName, releaseName)
	}

	return releaseJob, nil
}
