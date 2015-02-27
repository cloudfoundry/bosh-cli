package state

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmstatejob "github.com/cloudfoundry/bosh-micro-cli/state/job"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Builder interface {
	Build(bminstallmanifest.Manifest, bmui.Stage) (State, error)
}

type builder struct {
	releaseJobResolver    bmdeplrel.JobResolver
	jobDependencyCompiler bmstatejob.DependencyCompiler
	jobListRenderer       bmtemplate.JobListRenderer
	compressor            boshcmd.Compressor
	blobstore             boshblob.Blobstore
	templatesRepo         bmtemplate.TemplatesRepo
}

func NewBuilder(
	releaseJobResolver bmdeplrel.JobResolver,
	jobDependencyCompiler bmstatejob.DependencyCompiler,
	jobListRenderer bmtemplate.JobListRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo bmtemplate.TemplatesRepo,
) Builder {
	return &builder{
		releaseJobResolver:    releaseJobResolver,
		jobDependencyCompiler: jobDependencyCompiler,
		jobListRenderer:       jobListRenderer,
		compressor:            compressor,
		blobstore:             blobstore,
		templatesRepo:         templatesRepo,
	}
}

func (b *builder) Build(installationManifest bminstallmanifest.Manifest, stage bmui.Stage) (State, error) {
	// installation only ever has one job: the cpi job.
	releaseJobRefs := []bminstallmanifest.ReleaseJobRef{installationManifest.Template}

	// installation jobs do not get rendered with global deployment properties, only the cloud_provider properties
	globalProperties := bmproperty.Map{}
	jobProperties := installationManifest.Properties

	releaseJobs, err := b.resolveJobs(releaseJobRefs)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving jobs for installation")
	}

	compiledPackageRefs, err := b.jobDependencyCompiler.Compile(releaseJobs, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies for installation")
	}

	renderedJobRefs, err := b.renderJobTemplates(releaseJobs, jobProperties, globalProperties, installationManifest.Name, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Rendering job templates for installation")
	}

	if len(renderedJobRefs) != 1 {
		return nil, bosherr.Error("Too many jobs rendered... oops?")
	}

	compiledInstallationPackageRefs := make([]bminstallpkg.CompiledPackageRef, len(compiledPackageRefs), len(compiledPackageRefs))
	for i, compiledPackageRef := range compiledPackageRefs {
		compiledInstallationPackageRefs[i] = bminstallpkg.CompiledPackageRef{
			Name:        compiledPackageRef.Name,
			Version:     compiledPackageRef.Version,
			BlobstoreID: compiledPackageRef.BlobstoreID,
			SHA1:        compiledPackageRef.SHA1,
		}
	}

	return NewState(renderedJobRefs[0], compiledInstallationPackageRefs), nil
}

func (b *builder) resolveJobs(jobRefs []bminstallmanifest.ReleaseJobRef) ([]bmreljob.Job, error) {
	releaseJobs := make([]bmreljob.Job, len(jobRefs), len(jobRefs))
	for i, jobRef := range jobRefs {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.WrapErrorf(err, "Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
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
) ([]bminstalljob.RenderedJobRef, error) {
	renderedJobRefs := make([]bminstalljob.RenderedJobRef, 0, len(releaseJobs))
	err := stage.Perform("Rendering job templates", func() error {
		renderedJobList, err := b.jobListRenderer.Render(releaseJobs, jobProperties, globalProperties, deploymentName)
		if err != nil {
			return err
		}
		defer renderedJobList.DeleteSilently()

		for _, renderedJob := range renderedJobList.All() {
			renderedJobRecord, err := b.compressAndUpload(renderedJob)
			if err != nil {
				return err
			}

			releaseJob := renderedJob.Job()
			renderedJobRefs = append(renderedJobRefs, bminstalljob.RenderedJobRef{
				Name:        releaseJob.Name,
				Version:     releaseJob.Fingerprint,
				BlobstoreID: renderedJobRecord.BlobID,
				SHA1:        renderedJobRecord.BlobSHA1,
			})
		}

		return nil
	})

	return renderedJobRefs, err
}

func (b *builder) compressAndUpload(renderedJob bmtemplate.RenderedJob) (record bmtemplate.TemplateRecord, err error) {
	tarballPath, err := b.compressor.CompressFilesInDir(renderedJob.Path())
	if err != nil {
		return record, bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer b.compressor.CleanUp(tarballPath)

	blobID, blobSHA1, err := b.blobstore.Create(tarballPath)
	if err != nil {
		return record, bosherr.WrapError(err, "Creating blob")
	}

	record = bmtemplate.TemplateRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	//TODO: move TemplatesRepo to state/job.TemplatesRepo and reuse in deployment/instance/state.Builder
	err = b.templatesRepo.Save(renderedJob.Job(), record)
	if err != nil {
		return record, bosherr.WrapError(err, "Saving job to templates repo")
	}

	return record, nil
}
