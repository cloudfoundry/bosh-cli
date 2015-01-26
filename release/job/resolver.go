package job

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
)

type Resolver interface {
	Resolve(Reference) (bmrel.Job, error)
	ResolveEach([]Reference) ([]bmrel.Job, error)
}

type resolver struct {
	releaseSetResolver bmrelset.Resolver
}

func NewResolver(releaseSetResolver bmrelset.Resolver) Resolver {
	return &resolver{
		releaseSetResolver: releaseSetResolver,
	}
}

func (r *resolver) Resolve(jobRef Reference) (bmrel.Job, error) {
	release, err := r.releaseSetResolver.Find(jobRef.Release)
	if err != nil {
		return bmrel.Job{}, bosherr.WrapErrorf(err, "Resolving release '%s'", jobRef.Release)
	}

	releaseJob, found := release.FindJobByName(jobRef.Name)
	if !found {
		return bmrel.Job{}, bosherr.Errorf("Finding job '%s' in release '%s'", jobRef.Name, release.Name())
	}

	return releaseJob, nil
}

func (r *resolver) ResolveEach(jobRefs []Reference) ([]bmrel.Job, error) {
	releaseJobs := make([]bmrel.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		archive, err := r.Resolve(jobRef)
		if err != nil {
			return releaseJobs, err
		}
		releaseJobs[i] = archive
	}
	return releaseJobs, nil
}
