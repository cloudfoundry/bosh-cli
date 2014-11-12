package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/cmd/validation"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpideploy "github.com/cloudfoundry/bosh-micro-cli/cpideployer"
	bmdeployer "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deployCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	deploymentRecord        bmconfig.DeploymentRecord
	cpiManifestParser       bmdepl.ManifestParser
	boshManifestParser      bmdepl.ManifestParser
	boshDeploymentValidator bmdeplval.DeploymentValidator
	cpiDeployer             bmcpideploy.CpiDeployer
	stemcellManagerFactory  bmstemcell.ManagerFactory
	deployer                bmdeployer.Deployer
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	deploymentRecord bmconfig.DeploymentRecord,
	cpiManifestParser bmdepl.ManifestParser,
	boshManifestParser bmdepl.ManifestParser,
	boshDeploymentValidator bmdeplval.DeploymentValidator,
	cpiDeployer bmcpideploy.CpiDeployer,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	deployer bmdeployer.Deployer,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:                      ui,
		userConfig:              userConfig,
		fs:                      fs,
		deploymentRecord:        deploymentRecord,
		cpiManifestParser:       cpiManifestParser,
		boshManifestParser:      boshManifestParser,
		boshDeploymentValidator: boshDeploymentValidator,
		cpiDeployer:             cpiDeployer,
		stemcellManagerFactory:  stemcellManagerFactory,
		deployer:                deployer,
		eventLogger:             eventLogger,
		logger:                  logger,
		logTag:                  "deployCmd",
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

	validationStage := c.eventLogger.NewStage("validating")
	validationStage.Start()

	manifestValidationStep := validationStage.NewStep("Validating manifest")
	manifestValidationStep.Start()

	cpiDeployment, err := c.cpiManifestParser.Parse(c.userConfig.DeploymentFile)
	if err != nil {
		err = bosherr.WrapError(err, "Parsing CPI deployment manifest `%s'", c.userConfig.DeploymentFile)
		manifestValidationStep.Fail(err.Error())
		return err
	}

	boshDeployment, err := c.boshManifestParser.Parse(c.userConfig.DeploymentFile)
	if err != nil {
		err = bosherr.WrapError(err, "Parsing Bosh deployment manifest `%s'", c.userConfig.DeploymentFile)
		manifestValidationStep.Fail(err.Error())
		return err
	}

	err = c.boshDeploymentValidator.Validate(boshDeployment)
	if err != nil {
		err = bosherr.WrapError(err, "Validating bosh deployment manifest")
		manifestValidationStep.Fail(err.Error())
		return err
	}

	manifestValidationStep.Finish()
	validationStage.Finish()

	cloud, err := c.cpiDeployer.Deploy(cpiDeployment, releaseTarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Deploying CPI `%s'", releaseTarballPath)
	}

	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)
	stemcell, stemcellCID, err := stemcellManager.Upload(stemcellTarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell from `%s'", stemcellTarballPath)
	}

	err = c.deployer.Deploy(
		cloud,
		boshDeployment,
		stemcell.ApplySpec,
		cpiDeployment.Registry,
		cpiDeployment.SSHTunnel,
		cpiDeployment.Mbus,
		stemcellCID,
	)
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
		c.logger.Error(c.logTag, "Invalid arguments: ")
		return "", "", errors.New("Invalid usage - deploy command requires exactly 2 arguments")
	}

	// validate current state: 'microbosh' deployment set
	if len(c.userConfig.DeploymentFile) == 0 {
		c.ui.Error("No deployment set")
		return "", "", bosherr.New("No deployment set")
	}

	c.logger.Info(c.logTag, "Checking for deployment `%s'", c.userConfig.DeploymentFile)
	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(c.userConfig.DeploymentFile)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path `%s' does not exist", c.userConfig.DeploymentFile))
		return "", "", bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	releaseTarballPath := args[0]
	c.logger.Info(c.logTag, "Validating release tarball `%s'", releaseTarballPath)

	err = fileValidator.Exists(releaseTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' does not exist", releaseTarballPath))
		return "", "", bosherr.WrapError(err, "Checking CPI release `%s' existence", releaseTarballPath)
	}

	stemcellTarballPath := args[1]
	c.logger.Info(c.logTag, "Validating stemcell tarball `%s'", stemcellTarballPath)
	err = fileValidator.Exists(stemcellTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Stemcell `%s' does not exist", stemcellTarballPath))
		return "", "", bosherr.WrapError(err, "Checking stemcell `%s' existence", stemcellTarballPath)
	}

	return releaseTarballPath, stemcellTarballPath, nil
}
