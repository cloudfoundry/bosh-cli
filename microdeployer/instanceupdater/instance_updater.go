package instanceupdater

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/microdeployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/microdeployer/blobstore"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type instanceUpdater struct {
	agentClient       bmagentclient.AgentClient
	stemcellApplySpec bmstemcell.ApplySpec
	deployment        bmdepl.Deployment
	blobstore         bmblobstore.Blobstore
	compressor        boshcmd.Compressor
	erbrenderer       bmerbrenderer.ERBRenderer
	uuidGenerator     boshuuid.Generator
	applySpecFactory  bmas.Factory
	fs                boshsys.FileSystem
	logger            boshlog.Logger
	logTag            string
}

type InstanceUpdater interface {
	Update() error
	Start() error
}

func NewInstanceUpdater(
	agentClient bmagentclient.AgentClient,
	stemcellApplySpec bmstemcell.ApplySpec,
	deployment bmdepl.Deployment,
	blobstore bmblobstore.Blobstore,
	compressor boshcmd.Compressor,
	erbrenderer bmerbrenderer.ERBRenderer,
	uuidGenerator boshuuid.Generator,
	applySpecFactory bmas.Factory,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) InstanceUpdater {
	return &instanceUpdater{
		agentClient:       agentClient,
		stemcellApplySpec: stemcellApplySpec,
		deployment:        deployment,
		blobstore:         blobstore,
		compressor:        compressor,
		erbrenderer:       erbrenderer,
		uuidGenerator:     uuidGenerator,
		applySpecFactory:  applySpecFactory,
		fs:                fs,
		logger:            logger,
		logTag:            "instanceUpdater",
	}
}

func (u *instanceUpdater) Update() error {
	u.logger.Debug(u.logTag, "Stopping agent")

	err := u.agentClient.Stop()
	if err != nil {
		return bosherr.WrapError(err, "Stopping agent")
	}

	u.logger.Debug(u.logTag, "Rendering job templates")
	renderedJobDir, err := u.fs.TempDir("instance-updater-render-job")
	if err != nil {
		return bosherr.WrapError(err, "Creating rendered job directory")
	}
	defer u.fs.RemoveAll(renderedJobDir)

	job := u.deployment.Jobs[0]

	jobProperties, err := job.Properties()
	if err != nil {
		return bosherr.WrapError(err, "Stringifying job properties")
	}

	networksSpec, err := u.deployment.NetworksSpec(job.Name)
	if err != nil {
		return bosherr.WrapError(err, "Stringifying job properties")
	}

	for _, template := range job.Templates {
		for _, applySpecJobTemplate := range u.stemcellApplySpec.Job.Templates {
			if template.Name == applySpecJobTemplate.Name {
				tempFile, err := u.fs.TempFile("bosh-micro-job-template-blob")
				if err != nil {
					return bosherr.WrapError(err, "Creating tempfile")
				}
				defer u.fs.RemoveAll(tempFile.Name())

				err = u.blobstore.Get(applySpecJobTemplate.BlobstoreID, tempFile.Name())
				if err != nil {
					return bosherr.WrapError(err, "Creating tempfile")
				}

				renderedTemplateDir := filepath.Join(renderedJobDir, applySpecJobTemplate.Name)
				err = u.fs.MkdirAll(renderedTemplateDir, os.ModePerm)
				if err != nil {
					return bosherr.WrapError(err, "Creating rendered template directory")
				}

				err = u.renderJob(tempFile.Name(), renderedTemplateDir, jobProperties)
				if err != nil {
					return bosherr.WrapError(err, "Rendering job")
				}
			}
		}
	}

	u.logger.Debug(u.logTag, "Compressing job templates archive")
	renderedTarballPath, err := u.compressor.CompressFilesInDir(renderedJobDir)
	if err != nil {
		return bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer u.compressor.CleanUp(renderedTarballPath)

	blobID, err := u.uuidGenerator.Generate()
	if err != nil {
		return bosherr.WrapError(err, "Generating Blob ID")
	}

	u.logger.Debug(u.logTag, "Saving job templates archive to blobstore")
	err = u.blobstore.Save(renderedTarballPath, blobID)
	if err != nil {
		return bosherr.WrapError(err, "Uploading blob at %s", renderedTarballPath)
	}

	u.logger.Debug(u.logTag, "Creating apply spec")
	agentApplySpec, err := u.applySpecFactory.Create(
		u.stemcellApplySpec,
		u.deployment.Name,
		job.Name,
		networksSpec,
		blobID,
		renderedTarballPath,
		renderedJobDir,
	)

	u.logger.Debug(u.logTag, "Sending apply message to the agent with %#v", agentApplySpec)
	err = u.agentClient.Apply(agentApplySpec)
	if err != nil {
		return bosherr.WrapError(err, "Sending apply spec to agent")
	}

	return nil
}

func (u *instanceUpdater) Start() error {
	return u.agentClient.Start()
}

func (u *instanceUpdater) renderJob(
	blobPath string,
	renderedDir string,
	jobProperties map[string]interface{},
) error {
	jobExtractDir, err := u.fs.TempDir("instance-updater-extract-job")
	if err != nil {
		return bosherr.WrapError(err, "Creating job extraction directory")
	}

	defer u.fs.RemoveAll(jobExtractDir)

	jobReader := bmrel.NewJobReader(blobPath, jobExtractDir, u.compressor, u.fs)
	blobJob, err := jobReader.Read()
	if err != nil {
		return bosherr.WrapError(err, "Reading job from blob path %s", blobPath)
	}

	context := bmtempcomp.NewJobEvaluationContext(blobJob, jobProperties, u.deployment.Name, u.logger)

	for src, dst := range blobJob.Templates {
		err = u.renderFile(
			filepath.Join(jobExtractDir, "templates", src),
			filepath.Join(renderedDir, dst),
			context,
		)

		if err != nil {
			return bosherr.WrapError(err, "Rendering template src: %s, dst: %s", src, dst)
		}
	}

	err = u.renderFile(
		filepath.Join(jobExtractDir, "monit"),
		filepath.Join(renderedDir, "monit"),
		context,
	)
	if err != nil {
		return bosherr.WrapError(err, "Rendering monit file")
	}

	return nil
}

func (u *instanceUpdater) renderFile(sourcePath, destinationPath string, context bmerbrenderer.TemplateEvaluationContext) error {
	err := u.fs.MkdirAll(filepath.Dir(destinationPath), os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating tempdir '%s'", filepath.Dir(destinationPath))
	}

	err = u.erbrenderer.Render(sourcePath, destinationPath, context)
	if err != nil {
		return bosherr.WrapError(err, "Rendering template src: %s, dst: %s", sourcePath, destinationPath)
	}
	return nil
}
