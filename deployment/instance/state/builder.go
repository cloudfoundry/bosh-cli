package state

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmstatejob "github.com/cloudfoundry/bosh-micro-cli/state/job"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Builder interface {
	Build(jobName string, instanceID int, deploymentManifest bmdeplmanifest.Manifest, stage bmui.Stage) (State, error)
}

type builder struct {
	releaseJobResolver        bmdeplrel.JobResolver
	jobDependencyCompiler     bmstatejob.DependencyCompiler
	jobListRenderer           bmtemplate.JobListRenderer
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor
	blobstore                 bmblobstore.Blobstore
	logger                    boshlog.Logger
	logTag                    string
}

func NewBuilder(
	releaseJobResolver bmdeplrel.JobResolver,
	jobDependencyCompiler bmstatejob.DependencyCompiler,
	jobListRenderer bmtemplate.JobListRenderer,
	renderedJobListCompressor bmtemplate.RenderedJobListCompressor,
	blobstore bmblobstore.Blobstore,
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
	Archive     bmtemplate.RenderedJobListArchive
}

func (b *builder) Build(jobName string, instanceID int, deploymentManifest bmdeplmanifest.Manifest, stage bmui.Stage) (State, error) {
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

func (b *builder) resolveJobs(jobRefs []bmdeplmanifest.ReleaseJobRef) ([]bmreljob.Job, error) {
	releaseJobs := make([]bmreljob.Job, len(jobRefs), len(jobRefs))
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
	releaseJobs []bmreljob.Job,
	jobProperties bmproperty.Map,
	globalProperties bmproperty.Map,
	deploymentName string,
	stage bmui.Stage,
) (renderedJobs, error) {
	var (
		renderedJobListArchive bmtemplate.RenderedJobListArchive
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
