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
	Render(sourcePath string, destinationPath string, job bmrel.Job, properties map[string]interface{}, deploymentName string) error
}

type jobRenderer struct {
	erbRenderer bmerbrenderer.ERBRenderer
	fs          boshsys.FileSystem
	logger      boshlog.Logger
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
	}
}

func (r *jobRenderer) Render(sourcePath string, destinationPath string, job bmrel.Job, properties map[string]interface{}, deploymentName string) error {
	context := NewJobEvaluationContext(job, properties, deploymentName, r.logger)

	for src, dst := range job.Templates {
		err := r.renderFile(
			filepath.Join(sourcePath, "templates", src),
			filepath.Join(destinationPath, dst),
			context,
		)

		if err != nil {
			return bosherr.WrapErrorf(err, "Rendering template src: %s, dst: %s", src, dst)
		}
	}

	err := r.renderFile(
		filepath.Join(sourcePath, "monit"),
		filepath.Join(destinationPath, "monit"),
		context,
	)
	if err != nil {
		return bosherr.WrapError(err, "Rendering monit file")
	}

	return nil
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
