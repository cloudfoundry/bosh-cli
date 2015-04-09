package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelset "github.com/cloudfoundry/bosh-init/release/set"
)

type JobResolver interface {
	Resolve(jobName, releaseName string) (bireljob.Job, error)
}

type resolver struct {
	releaseSetResolver birelset.Resolver
}

func NewJobResolver(releaseSetResolver birelset.Resolver) JobResolver {
	return &resolver{
		releaseSetResolver: releaseSetResolver,
	}
}

func (r *resolver) Resolve(jobName, releaseName string) (bireljob.Job, error) {
	release, err := r.releaseSetResolver.Find(releaseName)
	if err != nil {
		return bireljob.Job{}, bosherr.WrapErrorf(err, "Resolving release '%s'", releaseName)
	}

	releaseJob, found := release.FindJobByName(jobName)
	if !found {
		return bireljob.Job{}, bosherr.Errorf("Finding job '%s' in release '%s'", jobName, releaseName)
	}

	return releaseJob, nil
}
