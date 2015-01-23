package templatescompiler

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
)

type JobRenderer interface {
	Render(job bmrel.Job, properties map[string]interface{}, deploymentName string) (RenderedJob, error)
}

type jobRenderer struct {
	erbRenderer bmerbrenderer.ERBRenderer
	fs          boshsys.FileSystem
	logger      boshlog.Logger
	logTag      string
}

func NewJobRenderer(
	erbRenderer bmerbrenderer.ERBRenderer,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) JobRenderer {
	return &jobRenderer{
		erbRenderer: erbRenderer,
		fs:          fs,
		logger:      logger,
		logTag:      "jobRenderer",
	}
}

//TODO: test me
func (r *jobRenderer) Render(job bmrel.Job, properties map[string]interface{}, deploymentName string) (RenderedJob, error) {
	context := NewJobEvaluationContext(job, properties, deploymentName, r.logger)

	sourcePath := job.ExtractedPath

	destinationPath, err := r.fs.TempDir("rendered-jobs")
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating rendered job directory")
	}

	renderedJob := NewRenderedJob(job, destinationPath, r.fs, r.logger)

	for src, dst := range job.Templates {
		err := r.renderFile(
			filepath.Join(sourcePath, "templates", src),
			filepath.Join(destinationPath, dst),
			context,
		)
		if err != nil {
			defer renderedJob.DeleteSilently()
			return nil, bosherr.WrapErrorf(err, "Rendering template src: %s, dst: %s", src, dst)
		}
	}

	err = r.renderFile(
		filepath.Join(sourcePath, "monit"),
		filepath.Join(destinationPath, "monit"),
		context,
	)
	if err != nil {
		defer renderedJob.DeleteSilently()
		return nil, bosherr.WrapError(err, "Rendering monit file")
	}

	return renderedJob, nil
}

func (r *jobRenderer) renderFile(sourcePath, destinationPath string, context bmerbrenderer.TemplateEvaluationContext) error {
	err := r.fs.MkdirAll(filepath.Dir(destinationPath), os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating tempdir '%s'", filepath.Dir(destinationPath))
	}

	err = r.erbRenderer.Render(sourcePath, destinationPath, context)
	if err != nil {
		return bosherr.WrapErrorf(err, "Rendering template src: %s, dst: %s", sourcePath, destinationPath)
	}
	return nil
}
