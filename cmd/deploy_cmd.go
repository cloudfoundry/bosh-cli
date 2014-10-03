package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeploy "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

const (
	logTag = "depoyCmd"
)

type deployCmd struct {
	ui                     bmui.UI
	userConfig             bmconfig.UserConfig
	fs                     boshsys.FileSystem
	cpiManifestParser      bmdepl.ManifestParser
	cpiDeployer            bmdeploy.CpiDeployer
	stemcellManagerFactory bmstemcell.ManagerFactory
	logger                 boshlog.Logger
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	cpiManifestParser bmdepl.ManifestParser,
	cpiDeployer bmdeploy.CpiDeployer,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:                     ui,
		userConfig:             userConfig,
		fs:                     fs,
		cpiManifestParser:      cpiManifestParser,
		cpiDeployer:            cpiDeployer,
		stemcellManagerFactory: stemcellManagerFactory,
		logger:                 logger,
	}
}

func (c *deployCmd) Name() string {
	return "deploy"
}

func (c *deployCmd) Run(args []string) error {
	releaseTarballPath, stemcellTarballPath, err := c.validateDeployInputs(args)
	if err != nil {
		return err
	}

	deployment, err := c.cpiManifestParser.Parse(c.userConfig.DeploymentFile)
	if err != nil {
		return bosherr.WrapError(err, "Parsing CPI deployment manifest `%s'", c.userConfig.DeploymentFile)
	}

	cloud, err := c.cpiDeployer.Deploy(deployment, releaseTarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Deploying CPI `%s'", releaseTarballPath)
	}

	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)
	stemcell, _, err := stemcellManager.Upload(stemcellTarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell from `%s'", stemcellTarballPath)
	}

	microboshDeployment, err := c.parseMicroboshManifest()
	if err != nil {
		return bosherr.WrapError(err, "Parsing Microbosh deployment manifest `%s'", c.userConfig.DeploymentFile)
	}

	err = c.deployMicrobosh(cloud, microboshDeployment, stemcell)
	if err != nil {
		return bosherr.WrapError(err, "Deploying Microbosh")
	}

	// register the stemcell
	return nil
}

type Deployment struct{}

// validateDeployInputs validates the presence of inputs (stemcell tarball, cpi release tarball)
func (c *deployCmd) validateDeployInputs(args []string) (string, string, error) {

	if len(args) != 2 {
		c.ui.Error("Invalid usage - deploy command requires exactly 2 arguments")
		c.ui.Sayln("Expected usage: bosh-micro deploy <cpi-release-tarball> <stemcell-tarball>")
		c.logger.Error(logTag, "Invalid arguments: ")
		return "", "", errors.New("Invalid usage - deploy command requires exactly 2 arguments")
	}

	releaseTarballPath := args[0]
	c.logger.Info(logTag, "Validating release tarball `%s'", releaseTarballPath)

	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(releaseTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' does not exist", releaseTarballPath))
		return "", "", bosherr.WrapError(err, "Checking CPI release `%s' existence", releaseTarballPath)
	}

	stemcellTarballPath := args[1]
	c.logger.Info(logTag, "Validating stemcell tarball `%s'", stemcellTarballPath)
	err = fileValidator.Exists(stemcellTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Stemcell `%s' does not exist", stemcellTarballPath))
		return "", "", bosherr.WrapError(err, "Checking stemcell `%s' existence", stemcellTarballPath)
	}

	// validate current state: 'microbosh' deployment set
	if len(c.userConfig.DeploymentFile) == 0 {
		c.ui.Error("No deployment set")
		return "", "", bosherr.New("No deployment set")
	}

	c.logger.Info(logTag, "Checking for deployment `%s'", c.userConfig.DeploymentFile)
	err = fileValidator.Exists(c.userConfig.DeploymentFile)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path `%s' does not exist", c.userConfig.DeploymentFile))
		return "", "", bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	return releaseTarballPath, stemcellTarballPath, nil
}

func (c *deployCmd) parseMicroboshManifest() (Deployment, error) {
	//c.userConfig.DeploymentFile
	return Deployment{}, nil
}

func (c *deployCmd) deployMicrobosh(cpi bmcloud.Cloud, deployment Deployment, stemcell bmstemcell.Stemcell) error {
	// create (or discover & update) remote deployment 'cells'
	//   cloud.create_vm & store agent_id
	//   wait for agent to bootstrap
	//   tell remote agent to apply state
	//   poll agent task get_state until finished
	return nil
}
