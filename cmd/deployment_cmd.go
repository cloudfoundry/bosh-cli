package cmd

import (
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
	fs                boshsys.FileSystem
	uuidGenerator     boshuuid.Generator
	logger            boshlog.Logger
	logTag            string
}

func NewDeploymentCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	userConfigService bmconfig.UserConfigService,
	fs boshsys.FileSystem,
	uuidGenerator boshuuid.Generator,
	logger boshlog.Logger,
) Cmd {
	return &deploymentCmd{
		ui:                ui,
		userConfig:        userConfig,
		userConfigService: userConfigService,
		fs:                fs,
		uuidGenerator:     uuidGenerator,
		logger:            logger,
		logTag:            "deploymentCmd",
	}
}

func (c *deploymentCmd) Name() string {
	return "deployment"
}

func (c *deploymentCmd) Run(stage bmui.Stage, args []string) error {
	if args == nil || len(args) < 1 {
		_, err := getDeploymentManifest(c.userConfig, c.ui, c.fs)
		if err != nil {
			return bosherr.WrapErrorf(err, "Running deployment cmd")
		}
		return nil
	}

	manifestFilePath := args[0]
	return c.setDeployment(manifestFilePath)
}

func (c *deploymentCmd) setDeployment(manifestFilePath string) error {
	manifestAbsFilePath, err := filepath.Abs(manifestFilePath)
	if err != nil {
		c.ui.ErrorLinef("Failed getting absolute path to deployment file '%s'", manifestFilePath)
		return bosherr.WrapErrorf(err, "Getting absolute path to deployment file '%s'", manifestFilePath)
	}

	if !c.fs.FileExists(manifestAbsFilePath) {
		c.ui.ErrorLinef("Deployment '%s' does not exist", manifestAbsFilePath)
		return bosherr.Errorf("Verifying that the deployment '%s' exists", manifestAbsFilePath)
	}

	c.userConfig.DeploymentManifestPath = manifestAbsFilePath
	err = c.userConfigService.Save(c.userConfig)
	if err != nil {
		return bosherr.WrapError(err, "Saving user config")
	}

	c.ui.PrintLinef("Deployment manifest set to '%s'", manifestAbsFilePath)

	deploymentConfigPath := c.userConfig.DeploymentConfigPath()
	deploymentConfigService := bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, c.fs, c.uuidGenerator, c.logger)

	// initialize defaults
	_, err = deploymentConfigService.Load()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Deployment state set to '%s'", deploymentConfigPath)

	return nil
}
