package installation

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	boshblob "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/fileutil"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	bitemplate "github.com/cloudfoundry/bosh-init/templatescompiler"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type JobRenderer interface {
	RenderAndUploadFrom(biinstallmanifest.Manifest, []bireljob.Job, biui.Stage) ([]biinstalljob.RenderedJobRef, error)
}

type jobRenderer struct {
	jobListRenderer bitemplate.JobListRenderer
	compressor      boshcmd.Compressor
	blobstore       boshblob.Blobstore
	templatesRepo   bitemplate.TemplatesRepo
}

func NewJobRenderer(
	jobListRenderer bitemplate.JobListRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo bitemplate.TemplatesRepo,
) JobRenderer {
	return &jobRenderer{
		jobListRenderer: jobListRenderer,
		compressor:      compressor,
		blobstore:       blobstore,
		templatesRepo:   templatesRepo,
	}
}

func (b *jobRenderer) RenderAndUploadFrom(installationManifest biinstallmanifest.Manifest, jobs []bireljob.Job, stage biui.Stage) ([]biinstalljob.RenderedJobRef, error) {
	// installation jobs do not get rendered with global deployment properties, only the cloud_provider properties
	globalProperties := biproperty.Map{}
	jobProperties := installationManifest.Properties

	renderedJobRefs, err := b.renderJobTemplates(jobs, jobProperties, globalProperties, installationManifest.Name, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Rendering job templates for installation")
	}

	if len(renderedJobRefs) != 1 {
		return nil, bosherr.Error("Too many jobs rendered... oops?")
	}

	return renderedJobRefs, nil
}

// renderJobTemplates renders all the release job templates for multiple release jobs specified
// by a deployment job and randomly uploads them to blobstore
func (b *jobRenderer) renderJobTemplates(
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

func (b *jobRenderer) compressAndUpload(renderedJob bitemplate.RenderedJob) (record bitemplate.TemplateRecord, err error) {
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
