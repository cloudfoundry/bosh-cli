package cmd

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmdeployer "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deployCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	deploymentParser        bmdepl.Parser
	boshDeploymentValidator bmdeplval.DeploymentValidator
	cpiInstaller            bmcpi.Installer
	stemcellExtractor       bmstemcell.Extractor
	deploymentRecord        bmdeployer.DeploymentRecord
	deployer                bmdeployer.Deployer
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	deploymentParser bmdepl.Parser,
	boshDeploymentValidator bmdeplval.DeploymentValidator,
	cpiInstaller bmcpi.Installer,
	stemcellExtractor bmstemcell.Extractor,
	deploymentRecord bmdeployer.DeploymentRecord,
	deployer bmdeployer.Deployer,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:                      ui,
		userConfig:              userConfig,
		fs:                      fs,
		deploymentParser:        deploymentParser,
		boshDeploymentValidator: boshDeploymentValidator,
		cpiInstaller:            cpiInstaller,
		stemcellExtractor:       stemcellExtractor,
		deploymentRecord:        deploymentRecord,
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
	releaseTarballPath, stemcellTarballPath, err := c.parseCmdInputs(args)
	if err != nil {
		return err
	}

	cpiDeployment, boshDeployment, cpiRelease, extractedStemcell, err := c.validateInputFiles(releaseTarballPath, stemcellTarballPath)
	if err != nil {
		return err
	}
	defer extractedStemcell.Delete()
	defer cpiRelease.Delete()

	isDeployed, err := c.deploymentRecord.IsDeployed(c.userConfig.DeploymentFile, cpiRelease, extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed {
		c.ui.Sayln("No deployment, stemcell or cpi release changes. Skipping deploy.")
		return nil
	}

	cloud, err := c.cpiInstaller.Install(cpiDeployment, cpiRelease)
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI deployment")
	}

	err = c.deployer.Deploy(
		cloud,
		boshDeployment,
		extractedStemcell,
		cpiDeployment.Registry,
		cpiDeployment.SSHTunnel,
		cpiDeployment.Mbus,
	)
	if err != nil {
		return bosherr.WrapError(err, "Deploying Microbosh")
	}

	err = c.deploymentRecord.Update(c.userConfig.DeploymentFile, cpiRelease)
	if err != nil {
		return bosherr.WrapError(err, "Updating deployment record")
	}

	return nil
}

type Deployment struct{}

func (c *deployCmd) validateInputFiles(releaseTarballPath, stemcellTarballPath string) (
	cpiDeployment bmdepl.CPIDeployment,
	boshDeployment bmdepl.Deployment,
	cpiRelease bmrel.Release,
	extractedStemcell bmstemcell.ExtractedStemcell,
	err error,
) {
	validationStage := c.eventLogger.NewStage("validating")
	validationStage.Start()

	err = validationStage.PerformStep("Validating deployment manifest", func() error {
		if c.userConfig.DeploymentFile == "" {
			return bosherr.Error("No deployment set")
		}

		deploymentFilePath := c.userConfig.DeploymentFile

		c.logger.Info(c.logTag, "Checking for deployment `%s'", deploymentFilePath)
		if !c.fs.FileExists(deploymentFilePath) {
			return bosherr.Errorf("Verifying that the deployment `%s' exists", deploymentFilePath)
		}

		boshDeployment, cpiDeployment, err = c.deploymentParser.Parse(deploymentFilePath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest `%s'", deploymentFilePath)
		}

		err = c.boshDeploymentValidator.Validate(boshDeployment)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		return nil
	})
	if err != nil {
		return cpiDeployment, boshDeployment, nil, nil, err
	}

	err = validationStage.PerformStep("Validating cpi release", func() error {
		if !c.fs.FileExists(releaseTarballPath) {
			return bosherr.Errorf("Verifying that the CPI release `%s' exists", releaseTarballPath)
		}

		cpiRelease, err = c.cpiInstaller.Extract(releaseTarballPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting CPI release `%s'", releaseTarballPath)
		}

		return nil
	})
	if err != nil {
		return cpiDeployment, boshDeployment, cpiRelease, nil, err
	}

	err = validationStage.PerformStep("Validating stemcell", func() error {
		if !c.fs.FileExists(stemcellTarballPath) {
			return bosherr.Errorf("Verifying that the stemcell `%s' exists", stemcellTarballPath)
		}

		extractedStemcell, err = c.stemcellExtractor.Extract(stemcellTarballPath)
		if err != nil {
			defer cpiRelease.Delete()
			return bosherr.WrapErrorf(err, "Extracting stemcell from `%s'", stemcellTarballPath)
		}

		return nil
	})
	if err != nil {
		return cpiDeployment, boshDeployment, cpiRelease, extractedStemcell, err
	}

	validationStage.Finish()

	return cpiDeployment, boshDeployment, cpiRelease, extractedStemcell, nil
}

func (c *deployCmd) parseCmdInputs(args []string) (string, string, error) {
	if len(args) != 2 {
		c.ui.Error("Invalid usage - deploy command requires exactly 2 arguments")
		c.ui.Sayln("Expected usage: bosh-micro deploy <cpi-release-tarball> <stemcell-tarball>")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", "", errors.New("Invalid usage - deploy command requires exactly 2 arguments")
	}
	return args[0], args[1], nil
}
