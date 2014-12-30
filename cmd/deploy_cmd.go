package cmd

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/validator"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deployCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	deploymentParser        bmmanifest.Parser
	deploymentConfigService bmconfig.DeploymentConfigService
	boshDeploymentValidator bmdeplval.DeploymentValidator
	cpiDeploymentFactory    bmcpi.DeploymentFactory
	agentClientFactory      bmhttpagent.AgentClientFactory
	vmManagerFactory        bmvm.ManagerFactory
	stemcellExtractor       bmstemcell.Extractor
	deploymentRecord        bmdepl.DeploymentRecord
	deploymentFactory       bmdepl.Factory
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	deploymentParser bmmanifest.Parser,
	deploymentConfigService bmconfig.DeploymentConfigService,
	boshDeploymentValidator bmdeplval.DeploymentValidator,
	cpiDeploymentFactory bmcpi.DeploymentFactory,
	agentClientFactory bmhttpagent.AgentClientFactory,
	vmManagerFactory bmvm.ManagerFactory,
	stemcellExtractor bmstemcell.Extractor,
	deploymentRecord bmdepl.DeploymentRecord,
	deploymentFactory bmdepl.Factory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:                      ui,
		userConfig:              userConfig,
		fs:                      fs,
		deploymentParser:        deploymentParser,
		deploymentConfigService: deploymentConfigService,
		boshDeploymentValidator: boshDeploymentValidator,
		cpiDeploymentFactory:    cpiDeploymentFactory,
		agentClientFactory:      agentClientFactory,
		vmManagerFactory:        vmManagerFactory,
		stemcellExtractor:       stemcellExtractor,
		deploymentRecord:        deploymentRecord,
		deploymentFactory:       deploymentFactory,
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

	deploymentManifestPath, err := getDeploymentManifest(c.userConfig, c.ui, c.fs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running deploy cmd")
	}

	validationStage := c.eventLogger.NewStage("validating")
	validationStage.Start()

	deploymentConfig, err := c.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}

	var (
		boshDeployment bmdepl.Deployment
		cpiDeployment  bmcpi.Deployment
	)

	err = validationStage.PerformStep("Validating deployment manifest", func() error {
		boshDeploymentManifest, cpiDeploymentManifest, err := c.deploymentParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", deploymentManifestPath)
		}

		cpiDeployment = c.cpiDeploymentFactory.NewDeployment(cpiDeploymentManifest, deploymentConfig.DeploymentID, deploymentConfig.DirectorID)

		err = c.boshDeploymentValidator.Validate(boshDeploymentManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		boshDeployment = c.deploymentFactory.NewDeployment(boshDeploymentManifest)

		return nil
	})
	if err != nil {
		return err
	}

	var (
		cpiRelease        bmrel.Release
		extractedStemcell bmstemcell.ExtractedStemcell
	)
	err = validationStage.PerformStep("Validating cpi release", func() error {
		if !c.fs.FileExists(releaseTarballPath) {
			return bosherr.Errorf("Verifying that the CPI release '%s' exists", releaseTarballPath)
		}

		cpiRelease, err = cpiDeployment.ExtractRelease(releaseTarballPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting CPI release '%s'", releaseTarballPath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = validationStage.PerformStep("Validating stemcell", func() error {
		if !c.fs.FileExists(stemcellTarballPath) {
			return bosherr.Errorf("Verifying that the stemcell '%s' exists", stemcellTarballPath)
		}

		extractedStemcell, err = c.stemcellExtractor.Extract(stemcellTarballPath)
		if err != nil {
			cpiRelease.Delete()
			return bosherr.WrapErrorf(err, "Extracting stemcell from '%s'", stemcellTarballPath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	validationStage.Finish()

	defer extractedStemcell.Delete()
	defer cpiRelease.Delete()

	isDeployed, err := c.deploymentRecord.IsDeployed(deploymentManifestPath, cpiRelease, extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed {
		c.ui.Sayln("No deployment, stemcell or cpi release changes. Skipping deploy.")
		return nil
	}

	cloud, err := cpiDeployment.Install()
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI deployment")
	}

	err = cpiDeployment.StartJobs()
	if err != nil {
		return bosherr.WrapError(err, "Starting CPI jobs")
	}
	defer func() {
		err := cpiDeployment.StopJobs()
		if err != nil {
			c.logger.Warn(c.logTag, "CPI jobs failed to stop: %s", err)
		}
	}()

	directorID := deploymentConfig.DirectorID
	cpiDeploymentManifest := cpiDeployment.Manifest()
	mbusURL := cpiDeploymentManifest.Mbus
	agentClient := c.agentClientFactory.NewAgentClient(directorID, mbusURL)
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient, mbusURL)

	err = boshDeployment.Deploy(
		cloud,
		extractedStemcell,
		cpiDeploymentManifest.Registry,
		cpiDeploymentManifest.SSHTunnel,
		vmManager,
	)
	if err != nil {
		return bosherr.WrapError(err, "Deploying Microbosh")
	}

	err = c.deploymentRecord.Update(deploymentManifestPath, cpiRelease)
	if err != nil {
		return bosherr.WrapError(err, "Updating deployment record")
	}

	return nil
}

type Deployment struct{}

func (c *deployCmd) parseCmdInputs(args []string) (string, string, error) {
	if len(args) != 2 {
		c.ui.Error("Invalid usage - deploy command requires exactly 2 arguments")
		c.ui.Sayln("Expected usage: bosh-micro deploy <cpi-release-tarball> <stemcell-tarball>")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", "", errors.New("Invalid usage - deploy command requires exactly 2 arguments")
	}
	return args[0], args[1], nil
}
