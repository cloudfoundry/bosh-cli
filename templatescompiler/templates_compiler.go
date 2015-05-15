package templatescompiler

import (
	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	biui "github.com/cloudfoundry/bosh-init/ui"
	boshblob "github.com/cloudfoundry/bosh-utils/blobstore"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type TemplatesCompiler interface {
	Compile(jobs []bireljob.Job, deploymentName string, deploymentProperties biproperty.Map, stage biui.Stage) error
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

func (tc templatesCompiler) Compile(releaseJobs []bireljob.Job, deploymentName string, jobProperties biproperty.Map, stage biui.Stage) error {
	// installation jobs do not get rendered with global deployment properties, only the cloud_provider properties
	globalProperties := biproperty.Map{}

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
