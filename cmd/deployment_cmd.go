package cmd

import (
	"errors"
	"fmt"

	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

const (
	BoshMicroFilename = ".bosh_micro.json"
)

type deploymentCmd struct {
	ui            bmui.UI
	config        bmconfig.Config
	configService bmconfig.Service
	fs            boshsys.FileSystem
}

func NewDeploymentCmd(ui bmui.UI, config bmconfig.Config, configService bmconfig.Service, fs boshsys.FileSystem) *deploymentCmd {
	return &deploymentCmd{
		ui:            ui,
		config:        config,
		configService: configService,
		fs:            fs,
	}
}

func (c *deploymentCmd) Run(args []string) error {
	if args == nil || len(args) < 1 {
		return c.showDeploymentStatus()
	}

	manifestFilePath := args[0]
	return c.setDeployment(manifestFilePath)
}

func (c *deploymentCmd) showDeploymentStatus() error {
	if c.config.Deployment == "" {
		c.ui.Error("Deployment not set")
		return errors.New("Deployment not set")
	}

	c.ui.Say(fmt.Sprintf("Current deployment is '%s'", c.config.Deployment))
	return nil
}

func (c *deploymentCmd) setDeployment(manifestFilePath string) error {
	if !c.fs.FileExists(manifestFilePath) {
		c.ui.Error(fmt.Sprintf("Deployment command manifest path %s does not exist", manifestFilePath))
		return fmt.Errorf("Deployment command manifest path %s does not exist", manifestFilePath)
	}

	c.config.Deployment = manifestFilePath
	c.configService.Save(c.config)
	c.ui.Say(fmt.Sprintf("Deployment set to '%s'", manifestFilePath))
	return nil
}
