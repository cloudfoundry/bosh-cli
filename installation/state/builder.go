package state

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	bideplrel "github.com/cloudfoundry/bosh-init/deployment/release"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	bistatejob "github.com/cloudfoundry/bosh-init/state/job"
	bitemplate "github.com/cloudfoundry/bosh-init/templatescompiler"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Builder interface {
	Build(biinstallmanifest.Manifest, biui.Stage) (State, error)
}

type builder struct {
	releaseJobResolver    bideplrel.JobResolver
	jobDependencyCompiler bistatejob.DependencyCompiler
	jobListRenderer       bitemplate.JobListRenderer
	compressor            boshcmd.Compressor
	blobstore             boshblob.Blobstore
	templatesRepo         bitemplate.TemplatesRepo
}

func NewBuilder(
	releaseJobResolver bideplrel.JobResolver,
	jobDependencyCompiler bistatejob.DependencyCompiler,
	jobListRenderer bitemplate.JobListRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo bitemplate.TemplatesRepo,
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

func (b *builder) Build(installationManifest biinstallmanifest.Manifest, stage biui.Stage) (State, error) {
	// installation only ever has one job: the cpi job.
	releaseJobRefs := []biinstallmanifest.ReleaseJobRef{installationManifest.Template}

	// installation jobs do not get rendered with global deployment properties, only the cloud_provider properties
	globalProperties := biproperty.Map{}
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

	compiledInstallationPackageRefs := make([]biinstallpkg.CompiledPackageRef, len(compiledPackageRefs), len(compiledPackageRefs))
	for i, compiledPackageRef := range compiledPackageRefs {
		compiledInstallationPackageRefs[i] = biinstallpkg.CompiledPackageRef{
			Name:        compiledPackageRef.Name,
			Version:     compiledPackageRef.Version,
			BlobstoreID: compiledPackageRef.BlobstoreID,
			SHA1:        compiledPackageRef.SHA1,
		}
	}

	return NewState(renderedJobRefs[0], compiledInstallationPackageRefs), nil
}

func (b *builder) resolveJobs(jobRefs []biinstallmanifest.ReleaseJobRef) ([]bireljob.Job, error) {
	releaseJobs := make([]bireljob.Job, len(jobRefs), len(jobRefs))
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
	releaseJobs []bireljob.Job,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	stage biui.Stage,
) ([]biinstalljob.RenderedJobRef, error) {
	renderedJobRefs := make([]biinstalljob.RenderedJobRef, 0, len(releaseJobs))
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
			renderedJobRefs = append(renderedJobRefs, biinstalljob.RenderedJobRef{
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

func (b *builder) compressAndUpload(renderedJob bitemplate.RenderedJob) (record bitemplate.TemplateRecord, err error) {
	tarballPath, err := b.compressor.CompressFilesInDir(renderedJob.Path())
	if err != nil {
		return record, bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer b.compressor.CleanUp(tarballPath)

	blobID, blobSHA1, err := b.blobstore.Create(tarballPath)
	if err != nil {
		return record, bosherr.WrapError(err, "Creating blob")
	}

	record = bitemplate.TemplateRecord{
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
