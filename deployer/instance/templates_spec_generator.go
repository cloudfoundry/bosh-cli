package instance

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/deployer/blobstore"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type TemplatesSpecGenerator interface {
	Create(deploymentJob bmdepl.Job, stemcellJob bmstemcell.Job, deploymentName string, properties map[string]interface{}, blobstoreURL string) (TemplatesSpec, error)
}

type templatesSpecGenerator struct {
	blobstoreFactory bmblobstore.Factory
	compressor       boshcmd.Compressor
	jobRenderer      bmtempcomp.JobRenderer
	uuidGenerator    boshuuid.Generator
	sha1Calculator   SHA1Calculator
	fs               boshsys.FileSystem
	logger           boshlog.Logger
	logTag           string
}

type TemplatesSpec struct {
	BlobID            string
	ArchiveSha1       string
	ConfigurationHash string
}

func NewTemplatesSpecGenerator(
	blobstoreFactory bmblobstore.Factory,
	compressor boshcmd.Compressor,
	jobRenderer bmtempcomp.JobRenderer,
	uuidGenerator boshuuid.Generator,
	sha1Calculator SHA1Calculator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) TemplatesSpecGenerator {
	return &templatesSpecGenerator{
		blobstoreFactory: blobstoreFactory,
		compressor:       compressor,
		jobRenderer:      jobRenderer,
		uuidGenerator:    uuidGenerator,
		sha1Calculator:   sha1Calculator,
		fs:               fs,
		logger:           logger,
		logTag:           "templatesSpecGenerator",
	}
}

func (g *templatesSpecGenerator) Create(
	deploymentJob bmdepl.Job,
	stemcellJob bmstemcell.Job,
	deploymentName string,
	properties map[string]interface{},
	blobstoreURL string,
) (TemplatesSpec, error) {
	g.logger.Debug(g.logTag, "Generating templates spec")
	renderedJobDir, err := g.fs.TempDir("instance-updater-render-job")
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Creating rendered job directory")
	}
	defer g.fs.RemoveAll(renderedJobDir)

	jobProperties, err := deploymentJob.Properties()
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Stringifying job properties")
	}

	blobstore, err := g.blobstoreFactory.Create(blobstoreURL)
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Creating blobstore client")
	}

	g.logger.Debug(g.logTag, "Rendering job templates")
	err = g.renderTemplates(deploymentJob.Templates, stemcellJob.Templates, blobstore, jobProperties, renderedJobDir, deploymentName)
	if err != nil {
		return TemplatesSpec{}, err
	}

	g.logger.Debug(g.logTag, "Compressing job templates archive")
	renderedTarballPath, err := g.compressor.CompressFilesInDir(renderedJobDir)
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer g.compressor.CleanUp(renderedTarballPath)

	blobID, err := g.uuidGenerator.Generate()
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Generating Blob ID")
	}

	g.logger.Debug(g.logTag, "Saving job templates archive to blobstore")
	err = blobstore.Save(renderedTarballPath, blobID)
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Uploading blob at '%s'", renderedTarballPath)
	}

	archivedTemplatesSha1, err := g.sha1Calculator.Calculate(renderedTarballPath)
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Calculating archived templates SHA1")
	}

	templatesDirSha1, err := g.sha1Calculator.Calculate(renderedJobDir)
	if err != nil {
		return TemplatesSpec{}, bosherr.WrapError(err, "Calculating templates dir SHA1")
	}

	return TemplatesSpec{
		BlobID:            blobID,
		ArchiveSha1:       archivedTemplatesSha1,
		ConfigurationHash: templatesDirSha1,
	}, nil
}

func (g *templatesSpecGenerator) renderTemplates(
	deploymentTemplates []bmdepl.ReleaseJobRef,
	stemcellTemplates []bmstemcell.Blob,
	blobstore bmblobstore.Blobstore,
	jobProperties map[string]interface{},
	renderedJobDir string,
	deploymentName string,
) error {
	for _, template := range deploymentTemplates {
		for _, applySpecJobTemplate := range stemcellTemplates {
			if template.Name == applySpecJobTemplate.Name {
				tempFile, err := g.fs.TempFile("bosh-micro-job-template-blob")
				if err != nil {
					return bosherr.WrapError(err, "Creating tempfile")
				}
				defer g.fs.RemoveAll(tempFile.Name())

				err = blobstore.Get(applySpecJobTemplate.BlobstoreID, tempFile.Name())
				if err != nil {
					return bosherr.WrapError(err, "Creating tempfile")
				}

				renderedTemplateDir := filepath.Join(renderedJobDir, applySpecJobTemplate.Name)
				err = g.fs.MkdirAll(renderedTemplateDir, os.ModePerm)
				if err != nil {
					return bosherr.WrapError(err, "Creating rendered template directory")
				}

				err = g.renderTemplate(tempFile.Name(), renderedTemplateDir, jobProperties, deploymentName)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (g *templatesSpecGenerator) renderTemplate(
	blobPath string,
	renderedDir string,
	jobProperties map[string]interface{},
	deploymentName string,
) error {
	jobExtractDir, err := g.fs.TempDir("instance-updater-extract-job")
	if err != nil {
		return bosherr.WrapError(err, "Creating job extraction directory")
	}
	defer g.fs.RemoveAll(jobExtractDir)

	jobReader := bmrel.NewJobReader(blobPath, jobExtractDir, g.compressor, g.fs)
	job, err := jobReader.Read()
	if err != nil {
		return bosherr.WrapError(err, "Reading job from blob path '%s'", blobPath)
	}

	err = g.jobRenderer.Render(jobExtractDir, renderedDir, job, jobProperties, deploymentName)
	if err != nil {
		return bosherr.WrapError(err, "Rendering job from blob path: '%s'", blobPath)
	}

	return nil
}
