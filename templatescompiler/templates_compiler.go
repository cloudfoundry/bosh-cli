package templatescompiler

import (
	"path/filepath"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type TemplatesCompiler interface {
	Compile(jobs []bmreljob.Job, deployment bmdepl.Deployment) error
}

type templatesCompiler struct {
	erbrenderer   bmerbrenderer.ERBRenderer
	compressor    boshcmd.Compressor
	blobstore     boshblob.Blobstore
	templatesRepo TemplatesRepo
	fs            boshsys.FileSystem
}

func NewTemplatesCompiler(
	erbrenderer bmerbrenderer.ERBRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo TemplatesRepo,
	fs boshsys.FileSystem,
) TemplatesCompiler {
	return templatesCompiler{
		erbrenderer:   erbrenderer,
		compressor:    compressor,
		blobstore:     blobstore,
		templatesRepo: templatesRepo,
		fs:            fs,
	}
}

func (tc templatesCompiler) Compile(jobs []bmreljob.Job, deployment bmdepl.Deployment) error {
	for _, job := range jobs {
		err := tc.compileJob(job, deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tc templatesCompiler) compileJob(job bmreljob.Job, deployment bmdepl.Deployment) error {
	jobSrcDir := job.ExtractedPath
	jobCompileDir, err := tc.fs.TempDir("templates-compiler")
	if err != nil {
		return bosherr.WrapError(err, "Creating compilation directory")
	}
	defer tc.fs.RemoveAll(jobCompileDir)

	context := NewJobEvaluationContext(job, deployment.Properties(), deployment.Name())

	for src, dst := range job.Templates {
		renderSrcPath := filepath.Join(jobSrcDir, src)
		renderDstPath := filepath.Join(jobCompileDir, dst)
		err = tc.erbrenderer.Render(renderSrcPath, renderDstPath, context)
		if err != nil {
			return bosherr.WrapError(err, "Rendering template src: %s, dst: %s", renderSrcPath, renderDstPath)
		}
	}

	tarball, err := tc.compressor.CompressFilesInDir(jobCompileDir)
	if err != nil {
		return bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer tc.compressor.CleanUp(tarball)

	blobID, blobSha1, err := tc.blobstore.Create(tarball)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	record := TemplateRecord{
		BlobID:   blobID,
		BlobSha1: blobSha1,
	}
	err = tc.templatesRepo.Save(job, record)
	if err != nil {
		return bosherr.WrapError(err, "Saving job to templates repo")
	}

	return nil
}
