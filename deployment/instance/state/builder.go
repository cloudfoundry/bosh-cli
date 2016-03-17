package state

import (
	biblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bideplrel "github.com/cloudfoundry/bosh-init/deployment/release"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/property"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	bistatejob "github.com/cloudfoundry/bosh-init/state/job"
	bitemplate "github.com/cloudfoundry/bosh-init/templatescompiler"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Builder interface {
	Build(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest, stage biui.Stage) (State, error)
}

type builder struct {
	releaseJobResolver        bideplrel.JobResolver
	jobDependencyCompiler     bistatejob.DependencyCompiler
	jobListRenderer           bitemplate.JobListRenderer
	renderedJobListCompressor bitemplate.RenderedJobListCompressor
	blobstore                 biblobstore.Blobstore
	logger                    boshlog.Logger
	logTag                    string
}

func NewBuilder(
	releaseJobResolver bideplrel.JobResolver,
	jobDependencyCompiler bistatejob.DependencyCompiler,
	jobListRenderer bitemplate.JobListRenderer,
	renderedJobListCompressor bitemplate.RenderedJobListCompressor,
	blobstore biblobstore.Blobstore,
	logger boshlog.Logger,
) Builder {
	return &builder{
		releaseJobResolver:        releaseJobResolver,
		jobDependencyCompiler:     jobDependencyCompiler,
		jobListRenderer:           jobListRenderer,
		renderedJobListCompressor: renderedJobListCompressor,
		blobstore:                 blobstore,
		logger:                    logger,
		logTag:                    "instanceStateBuilder",
	}
}

type renderedJobs struct {
	BlobstoreID string
	Archive     bitemplate.RenderedJobListArchive
}

func (b *builder) Build(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest, stage biui.Stage) (State, error) {
	deploymentJob, found := deploymentManifest.FindJobByName(jobName)
	if !found {
		return nil, bosherr.Errorf("Job '%s' not found in deployment manifest", jobName)
	}

	releaseJobs, err := b.resolveJobs(deploymentJob.Templates)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Resolving jobs for instance '%s/%d'", jobName, instanceID)
	}

	renderedJobTemplates, err := b.renderJobTemplates(releaseJobs, deploymentJob.Properties, deploymentManifest.Properties, deploymentManifest.Name, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Rendering job templates for instance '%s/%d'", jobName, instanceID)
	}

	compiledPackageRefs, err := b.jobDependencyCompiler.Compile(releaseJobs, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Compiling job package dependencies for instance '%s/%d'", jobName, instanceID)
	}

	networkInterfaces, err := deploymentManifest.NetworkInterfaces(deploymentJob.Name)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding networks for job '%s", jobName)
	}

	// convert map to array
	networkRefs := make([]NetworkRef, 0, len(networkInterfaces))
	for networkName, networkInterface := range networkInterfaces {
		networkRefs = append(networkRefs, NetworkRef{
			Name:      networkName,
			Interface: networkInterface,
		})
	}

	compiledDeploymentPackageRefs := make([]PackageRef, len(compiledPackageRefs), len(compiledPackageRefs))
	for i, compiledPackageRef := range compiledPackageRefs {
		compiledDeploymentPackageRefs[i] = PackageRef{
			Name:    compiledPackageRef.Name,
			Version: compiledPackageRef.Version,
			Archive: BlobRef{
				BlobstoreID: compiledPackageRef.BlobstoreID,
				SHA1:        compiledPackageRef.SHA1,
			},
		}
	}

	// convert array to array
	renderedJobRefs := make([]JobRef, len(releaseJobs), len(releaseJobs))
	for i, releaseJob := range releaseJobs {
		renderedJobRefs[i] = JobRef{
			Name:    releaseJob.Name,
			Version: releaseJob.Fingerprint,
		}
	}

	renderedJobListArchiveBlobRef := BlobRef{
		BlobstoreID: renderedJobTemplates.BlobstoreID,
		SHA1:        renderedJobTemplates.Archive.SHA1(),
	}

	return &state{
		deploymentName:         deploymentManifest.Name,
		name:                   jobName,
		id:                     instanceID,
		networks:               networkRefs,
		compiledPackages:       compiledDeploymentPackageRefs,
		renderedJobs:           renderedJobRefs,
		renderedJobListArchive: renderedJobListArchiveBlobRef,
		hash: renderedJobTemplates.Archive.Fingerprint(),
	}, nil
}

// FIXME: why do i exist here and in installation/state/builder.go??
func (b *builder) resolveJobs(jobRefs []bideplmanifest.ReleaseJobRef) ([]bireljob.Job, error) {
	releaseJobs := make([]bireljob.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.Errorf("Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}

// renderJobTemplates renders all the release job templates for multiple release jobs specified by a deployment job
func (b *builder) renderJobTemplates(
	releaseJobs []bireljob.Job,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	stage biui.Stage,
) (renderedJobs, error) {
	var (
		renderedJobListArchive bitemplate.RenderedJobListArchive
		blobID                 string
	)
	err := stage.Perform("Rendering job templates", func() error {
		renderedJobList, err := b.jobListRenderer.Render(releaseJobs, jobProperties, globalProperties, deploymentName)
		if err != nil {
			return err
		}
		defer renderedJobList.DeleteSilently()

		renderedJobListArchive, err = b.renderedJobListCompressor.Compress(renderedJobList)
		if err != nil {
			return bosherr.WrapError(err, "Compressing rendered job templates")
		}
		defer renderedJobListArchive.DeleteSilently()

		blobID, err = b.blobstore.Add(renderedJobListArchive.Path())
		if err != nil {
			return bosherr.WrapErrorf(err, "Uploading rendered job template archive '%s' to the blobstore", renderedJobListArchive.Path())
		}

		return nil
	})
	if err != nil {
		return renderedJobs{}, err
	}

	return renderedJobs{
		BlobstoreID: blobID,
		Archive:     renderedJobListArchive,
	}, nil
}
