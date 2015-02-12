package templatescompiler

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type TemplatesCompiler interface {
	Compile(jobs []bmreljob.Job, deploymentName string, deploymentProperties bmproperty.Map, stage bmui.Stage) error
}

type templatesCompiler struct {
	jobListRenderer JobListRenderer
	compressor      boshcmd.Compressor
	blobstore       boshblob.Blobstore
	templatesRepo   TemplatesRepo
	fs              boshsys.FileSystem
	logger          boshlog.Logger
}

func NewTemplatesCompiler(
	jobListRenderer JobListRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo TemplatesRepo,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) TemplatesCompiler {
	return templatesCompiler{
		jobListRenderer: jobListRenderer,
		compressor:      compressor,
		blobstore:       blobstore,
		templatesRepo:   templatesRepo,
		fs:              fs,
		logger:          logger,
	}
}

func (tc templatesCompiler) Compile(releaseJobs []bmreljob.Job, deploymentName string, jobProperties bmproperty.Map, stage bmui.Stage) error {
	// installation jobs do not get rendered with global deployment properties, only the cloud_provider properties
	globalProperties := bmproperty.Map{}

	return stage.Perform("Rendering job templates", func() error {
		renderedJobList, err := tc.jobListRenderer.Render(releaseJobs, jobProperties, globalProperties, deploymentName)
		if err != nil {
			return err
		}
		defer renderedJobList.DeleteSilently()

		for _, renderedJob := range renderedJobList.All() {
			err := tc.compressAndUpload(renderedJob)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (tc templatesCompiler) compressAndUpload(renderedJob RenderedJob) error {
	tarballPath, err := tc.compressor.CompressFilesInDir(renderedJob.Path())
	if err != nil {
		return bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer tc.compressor.CleanUp(tarballPath)

	blobID, blobSHA1, err := tc.blobstore.Create(tarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	record := TemplateRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	err = tc.templatesRepo.Save(renderedJob.Job(), record)
	if err != nil {
		return bosherr.WrapError(err, "Saving job to templates repo")
	}

	return nil
}
