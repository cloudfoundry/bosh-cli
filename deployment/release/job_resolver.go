package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
)

type JobResolver interface {
	Resolve(bmdeplmanifest.ReleaseJobRef) (bmrel.Job, error)
	ResolveEach([]bmdeplmanifest.ReleaseJobRef) ([]bmrel.Job, error)
}

type resolver struct {
	releaseSetResolver bmrelset.Resolver
}

func NewJobResolver(releaseSetResolver bmrelset.Resolver) JobResolver {
	return &resolver{
		releaseSetResolver: releaseSetResolver,
	}
}

func (r *resolver) Resolve(jobRef bmdeplmanifest.ReleaseJobRef) (bmrel.Job, error) {
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

func (r *resolver) ResolveEach(jobRefs []bmdeplmanifest.ReleaseJobRef) ([]bmrel.Job, error) {
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
