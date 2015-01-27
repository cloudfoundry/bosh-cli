package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
)

type JobResolver interface {
	Resolve(jobName, releaseName string) (bmrel.Job, error)
}

type resolver struct {
	releaseSetResolver bmrelset.Resolver
}

func NewJobResolver(releaseSetResolver bmrelset.Resolver) JobResolver {
	return &resolver{
		releaseSetResolver: releaseSetResolver,
	}
}

func (r *resolver) Resolve(jobName, releaseName string) (bmrel.Job, error) {
	release, err := r.releaseSetResolver.Find(releaseName)
	if err != nil {
		return bmrel.Job{}, bosherr.WrapErrorf(err, "Resolving release '%s'", releaseName)
	}

	releaseJob, found := release.FindJobByName(jobName)
	if !found {
		return bmrel.Job{}, bosherr.Errorf("Finding job '%s' in release '%s'", jobName, releaseName)
	}

	return releaseJob, nil
}
