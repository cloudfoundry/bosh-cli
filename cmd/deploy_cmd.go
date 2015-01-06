package cmd

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmcpirel "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/validator"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deployCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	installationParser      bminstallmanifest.Parser
	deploymentParser        bmdeplmanifest.Parser
	deploymentConfigService bmconfig.DeploymentConfigService
	boshDeploymentValidator bmdeplval.DeploymentValidator
	installationFactory     bmcpi.InstallationFactory
	releaseManager          bmrel.Manager
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
	installationParser bminstallmanifest.Parser,
	deploymentParser bmdeplmanifest.Parser,
	deploymentConfigService bmconfig.DeploymentConfigService,
	boshDeploymentValidator bmdeplval.DeploymentValidator,
	installationFactory bmcpi.InstallationFactory,
	releaseManager bmrel.Manager,
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
		installationParser:      installationParser,
		deploymentParser:        deploymentParser,
		deploymentConfigService: deploymentConfigService,
		boshDeploymentValidator: boshDeploymentValidator,
		installationFactory:     installationFactory,
		releaseManager:          releaseManager,
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
	stemcellTarballPath, releaseTarballPaths, err := c.parseCmdInputs(args)
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
		installation   bmcpi.Installation
	)

	err = validationStage.PerformStep("Validating deployment manifest", func() error {
		installationManifest, err := c.installationParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
		}

		deploymentManifest, err := c.deploymentParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", deploymentManifestPath)
		}

		installation = c.installationFactory.NewInstallation(installationManifest, deploymentConfig.DeploymentID, deploymentConfig.DirectorID)

		err = c.boshDeploymentValidator.Validate(deploymentManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		boshDeployment = c.deploymentFactory.NewDeployment(deploymentManifest)

		return nil
	})
	if err != nil {
		return err
	}

	var extractedStemcell bmstemcell.ExtractedStemcell
	err = validationStage.PerformStep("Validating stemcell", func() error {
		if !c.fs.FileExists(stemcellTarballPath) {
			return bosherr.Errorf("Verifying that the stemcell '%s' exists", stemcellTarballPath)
		}

		extractedStemcell, err = c.stemcellExtractor.Extract(stemcellTarballPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting stemcell from '%s'", stemcellTarballPath)
		}

		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		deleteErr := extractedStemcell.Delete()
		if deleteErr != nil {
			c.logger.Warn(c.logTag, "Failed to delete extracted stemcell: %s", deleteErr.Error())
		}
	}()

	var cpiRelease bmrel.Release
	err = validationStage.PerformStep("Validating releases", func() error {
		for _, releaseTarballPath := range releaseTarballPaths {
			if !c.fs.FileExists(releaseTarballPath) {
				return bosherr.Errorf("Verifying that the release '%s' exists", releaseTarballPath)
			}

			_, err := c.releaseManager.Extract(releaseTarballPath)
			if err != nil {
				return bosherr.WrapErrorf(err, "Extracting release '%s'", releaseTarballPath)
			}
		}

		var found bool
		cpiRelease, found = bmcpirel.FindCPIRelease(c.releaseManager.List())
		if !found {
			return bosherr.Errorf("No provided release contains the required '%s' job", bmcpirel.ReleaseJobName)
		}

		err := bmcpirel.NewCpiValidator().Validate(cpiRelease)
		if err != nil {
			return bosherr.WrapError(err, "Invalid CPI release")
		}

		return nil
	})
	if err != nil {
		return err
	}
	defer func() {
		err := c.releaseManager.DeleteAll()
		if err != nil {
			c.logger.Warn(c.logTag, "Deleting all extracted releases: %s", err.Error())
		}
	}()

	validationStage.Finish()

	isDeployed, err := c.deploymentRecord.IsDeployed(deploymentManifestPath, cpiRelease, extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed {
		c.ui.Sayln("No deployment, stemcell or cpi release changes. Skipping deploy.")
		return nil
	}

	cloud, err := installation.Install()
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI deployment")
	}

	err = installation.StartJobs()
	if err != nil {
		return bosherr.WrapError(err, "Starting CPI jobs")
	}
	defer func() {
		err := installation.StopJobs()
		if err != nil {
			c.logger.Warn(c.logTag, "CPI jobs failed to stop: %s", err)
		}
	}()

	directorID := deploymentConfig.DirectorID
	installationManifest := installation.Manifest()
	mbusURL := installationManifest.Mbus
	agentClient := c.agentClientFactory.NewAgentClient(directorID, mbusURL)
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient, mbusURL)

	err = boshDeployment.Deploy(
		cloud,
		extractedStemcell,
		installationManifest.Registry,
		installationManifest.SSHTunnel,
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

func (c *deployCmd) parseCmdInputs(args []string) (string, []string, error) {
	if len(args) < 2 {
		c.ui.Error("Invalid usage - deploy command requires at least 2 arguments")
		c.ui.Sayln("Expected usage: bosh-micro deploy <stemcell-tarball> <cpi-release-tarball> [release-2-tarball [release-3-tarball...]]")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", []string{}, errors.New("Invalid usage - deploy command requires at least 2 arguments")
	}
	return args[0], args[1:], nil
}
