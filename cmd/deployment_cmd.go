package cmd

import (
	"errors"
	"fmt"
	"os"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

const (
	tagString = "DeploymentCmd"
)

type deploymentCmd struct {
	ui                bmui.UI
	userConfig        bmconfig.UserConfig
	userConfigService bmconfig.UserConfigService
	deploymentConfig  bmconfig.DeploymentConfig
	fs                boshsys.FileSystem
	uuidGenerator     boshuuid.Generator
	logger            boshlog.Logger
}

func NewDeploymentCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	userConfigService bmconfig.UserConfigService,
	deploymentConfig bmconfig.DeploymentConfig,
	fs boshsys.FileSystem,
	uuidGenerator boshuuid.Generator,
	logger boshlog.Logger,
) *deploymentCmd {
	return &deploymentCmd{
		ui:                ui,
		userConfig:        userConfig,
		userConfigService: userConfigService,
		deploymentConfig:  deploymentConfig,
		fs:                fs,
		uuidGenerator:     uuidGenerator,
		logger:            logger,
	}
}

func (c *deploymentCmd) Name() string {
	return "deployment"
}

func (c *deploymentCmd) Run(args []string) error {
	if args == nil || len(args) < 1 {
		return c.showDeploymentStatus()
	}

	manifestFilePath := args[0]
	return c.setDeployment(manifestFilePath)
}

func (c *deploymentCmd) showDeploymentStatus() error {
	if c.userConfig.DeploymentFile == "" {
		c.ui.Error("No deployment set")
		return errors.New("No deployment set")
	}

	c.ui.Sayln(fmt.Sprintf("Current deployment is `%s'", c.userConfig.DeploymentFile))
	return nil
}

func (c *deploymentCmd) setDeployment(manifestFilePath string) error {
	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(manifestFilePath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment `%s' does not exist", manifestFilePath))
		return bosherr.WrapError(err, "Setting deployment manifest")
	}

	c.userConfig.DeploymentFile = manifestFilePath
	err = c.userConfigService.Save(c.userConfig)
	if err != nil {
		return bosherr.WrapError(err, "Saving user config")
	}

	uuid, err := c.uuidGenerator.Generate()
	if err != nil {
		return bosherr.WrapError(err, "UUID Generation failed")
	}

	c.deploymentConfig.DeploymentUUID = uuid
	c.logger.Debug(tagString, "Config %#v", c.deploymentConfig)

	deploymentConfigService := bmconfig.NewFileSystemDeploymentConfigService(
		c.userConfig.DeploymentConfigFilePath(),
		c.fs,
		c.logger,
	)
	err = deploymentConfigService.Save(c.deploymentConfig)
	if err != nil {
		return bosherr.WrapError(err, "Saving deployment config")
	}

	blobstoreDir := c.deploymentConfig.BlobstorePath()
	c.logger.Debug(tagString, "Making new blobstore directory `%s'", blobstoreDir)
	err = c.fs.MkdirAll(blobstoreDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating blobs dir")
	}

	c.ui.Sayln(fmt.Sprintf("Deployment set to `%s'", manifestFilePath))
	return nil
}
