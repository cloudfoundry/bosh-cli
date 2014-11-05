package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/cmd/validation"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deploymentCmd struct {
	ui                bmui.UI
	userConfig        bmconfig.UserConfig
	userConfigService bmconfig.UserConfigService
	deploymentConfig  bmconfig.DeploymentConfig
	fs                boshsys.FileSystem
	uuidGenerator     boshuuid.Generator
	logger            boshlog.Logger
	logTag            string
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
		logTag:            "deploymentCmd",
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

	deploymentConfigService := bmconfig.NewFileSystemDeploymentConfigService(
		c.userConfig.DeploymentConfigFilePath(),
		c.fs,
		c.logger,
	)
	c.deploymentConfig, err = deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Reading existing deployment config")
	}

	if c.deploymentConfig.DeploymentUUID == "" {
		uuid, err := c.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "UUID Generation failed")
		}
		c.deploymentConfig.DeploymentUUID = uuid

		c.logger.Debug(c.logTag, "Config %#v", c.deploymentConfig)
		err = deploymentConfigService.Save(c.deploymentConfig)
		if err != nil {
			return bosherr.WrapError(err, "Saving deployment config")
		}
	}

	c.ui.Sayln(fmt.Sprintf("Deployment set to `%s'", manifestFilePath))
	return nil
}
