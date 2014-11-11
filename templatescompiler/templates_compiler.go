package templatescompiler

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type TemplatesCompiler interface {
	Compile(jobs []bmrel.Job, deployment bmdepl.Deployment) error
}

type templatesCompiler struct {
	jobRenderer   JobRenderer
	compressor    boshcmd.Compressor
	blobstore     boshblob.Blobstore
	templatesRepo TemplatesRepo
	fs            boshsys.FileSystem
	logger        boshlog.Logger
}

func NewTemplatesCompiler(
	jobRenderer JobRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo TemplatesRepo,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) TemplatesCompiler {
	return templatesCompiler{
		jobRenderer:   jobRenderer,
		compressor:    compressor,
		blobstore:     blobstore,
		templatesRepo: templatesRepo,
		fs:            fs,
		logger:        logger,
	}
}

func (tc templatesCompiler) Compile(jobs []bmrel.Job, deployment bmdepl.Deployment) error {
	for _, job := range jobs {
		err := tc.compileJob(job, deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tc templatesCompiler) compileJob(job bmrel.Job, deployment bmdepl.Deployment) error {
	jobSrcDir := job.ExtractedPath
	jobCompileDir, err := tc.fs.TempDir("templates-compiler")
	if err != nil {
		return bosherr.WrapError(err, "Creating compilation directory")
	}
	defer tc.fs.RemoveAll(jobCompileDir)

	properties, err := deployment.Properties()
	if err != nil {
		return bosherr.WrapError(err, "Getting deployment properties")
	}

	err = tc.jobRenderer.Render(jobSrcDir, jobCompileDir, job, properties, deployment.Name)
	if err != nil {
		return bosherr.WrapError(err, "Rendering templates")
	}

	tarball, err := tc.compressor.CompressFilesInDir(jobCompileDir)
	if err != nil {
		return bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer tc.compressor.CleanUp(tarball)

	blobID, blobSHA1, err := tc.blobstore.Create(tarball)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	record := TemplateRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	err = tc.templatesRepo.Save(job, record)
	if err != nil {
		return bosherr.WrapError(err, "Saving job to templates repo")
	}

	return nil
}
