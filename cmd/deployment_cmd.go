package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deploymentCmd struct {
	ui                bmui.UI
	userConfig        bmconfig.UserConfig
	userConfigService bmconfig.UserConfigService
	deploymentFile    bmconfig.DeploymentFile
	fs                boshsys.FileSystem
	uuidGenerator     boshuuid.Generator
	logger            boshlog.Logger
	logTag            string
}

func NewDeploymentCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	userConfigService bmconfig.UserConfigService,
	deploymentFile bmconfig.DeploymentFile,
	fs boshsys.FileSystem,
	uuidGenerator boshuuid.Generator,
	logger boshlog.Logger,
) *deploymentCmd {
	return &deploymentCmd{
		ui:                ui,
		userConfig:        userConfig,
		userConfigService: userConfigService,
		deploymentFile:    deploymentFile,
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

	c.ui.Sayln(fmt.Sprintf("Current deployment is '%s'", c.userConfig.DeploymentFile))
	return nil
}

func (c *deploymentCmd) setDeployment(manifestFilePath string) error {
	manifestAbsFilePath, err := filepath.Abs(manifestFilePath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed getting absolute path to deployment file '%s'", manifestFilePath))
		return bosherr.WrapErrorf(err, "Getting absolute path to deployment file '%s'", manifestFilePath)
	}

	if !c.fs.FileExists(manifestAbsFilePath) {
		c.ui.Error(fmt.Sprintf("Deployment '%s' does not exist", manifestAbsFilePath))
		return bosherr.Errorf("Verifying that the deployment '%s' exists", manifestAbsFilePath)
	}

	c.userConfig.DeploymentFile = manifestAbsFilePath
	err = c.userConfigService.Save(c.userConfig)
	if err != nil {
		return bosherr.WrapError(err, "Saving user config")
	}

	deploymentConfigService := bmconfig.NewFileSystemDeploymentConfigService(
		c.userConfig.DeploymentConfigFilePath(),
		c.fs,
		c.logger,
	)
	c.deploymentFile, err = deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Reading existing deployment config")
	}

	if c.deploymentFile.UUID == "" {
		uuid, err := c.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "UUID Generation failed")
		}
		c.deploymentFile.UUID = uuid

		c.logger.Debug(c.logTag, "Config %#v", c.deploymentFile)
		err = deploymentConfigService.Save(c.deploymentFile)
		if err != nil {
			return bosherr.WrapError(err, "Saving deployment config")
		}
	}

	c.ui.Sayln(fmt.Sprintf("Deployment set to '%s'", manifestAbsFilePath))
	return nil
}
