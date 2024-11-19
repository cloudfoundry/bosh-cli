package installation

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bideplrel "github.com/cloudfoundry/bosh-cli/v7/deployment/release"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
)

type JobResolver interface {
	From(biinstallmanifest.Manifest) ([]bireljob.Job, error)
}

type jobResolver struct {
	releaseJobResolver bideplrel.JobResolver
}

func NewJobResolver(
	releaseJobResolver bideplrel.JobResolver,
) JobResolver {
	return &jobResolver{
		releaseJobResolver: releaseJobResolver,
	}
}

func (b *jobResolver) From(installationManifest biinstallmanifest.Manifest) ([]bireljob.Job, error) {
	jobsReferencesInRelease := []biinstallmanifest.ReleaseJobRef{}
	for _, template := range installationManifest.Templates {
		jobsReferencesInRelease = append(jobsReferencesInRelease, biinstallmanifest.ReleaseJobRef{Name: template.Name, Release: template.Release})
	}

	releaseJobs := make([]bireljob.Job, len(jobsReferencesInRelease))
	for i, jobRef := range jobsReferencesInRelease {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.WrapErrorf(err, "Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}
