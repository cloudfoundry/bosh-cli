package cmd

import (
	"errors"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpirel "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deployCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	releaseSetParser        bmrelsetmanifest.Parser
	installationParser      bminstallmanifest.Parser
	deploymentParser        bmdeplmanifest.Parser
	deploymentConfigService bmconfig.DeploymentConfigService
	releaseSetValidator     bmrelsetmanifest.Validator
	installationValidator   bminstallmanifest.Validator
	deploymentValidator     bmdeplmanifest.Validator
	installerFactory        bminstall.InstallerFactory
	releaseExtractor        bmrel.Extractor
	releaseManager          bmrel.Manager
	releaseResolver         bmrelset.Resolver
	cloudFactory            bmcloud.Factory
	agentClientFactory      bmhttpagent.AgentClientFactory
	vmManagerFactory        bmvm.ManagerFactory
	stemcellExtractor       bmstemcell.Extractor
	stemcellManagerFactory  bmstemcell.ManagerFactory
	deploymentRecord        bmdepl.Record
	blobstoreFactory        bmblobstore.Factory
	deployer                bmdepl.Deployer
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	releaseSetParser bmrelsetmanifest.Parser,
	installationParser bminstallmanifest.Parser,
	deploymentParser bmdeplmanifest.Parser,
	deploymentConfigService bmconfig.DeploymentConfigService,
	releaseSetValidator bmrelsetmanifest.Validator,
	installationValidator bminstallmanifest.Validator,
	deploymentValidator bmdeplmanifest.Validator,
	installerFactory bminstall.InstallerFactory,
	releaseExtractor bmrel.Extractor,
	releaseManager bmrel.Manager,
	releaseResolver bmrelset.Resolver,
	cloudFactory bmcloud.Factory,
	agentClientFactory bmhttpagent.AgentClientFactory,
	vmManagerFactory bmvm.ManagerFactory,
	stemcellExtractor bmstemcell.Extractor,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	deploymentRecord bmdepl.Record,
	blobstoreFactory bmblobstore.Factory,
	deployer bmdepl.Deployer,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) Cmd {
	return &deployCmd{
		ui:                      ui,
		userConfig:              userConfig,
		fs:                      fs,
		releaseSetParser:        releaseSetParser,
		installationParser:      installationParser,
		deploymentParser:        deploymentParser,
		deploymentConfigService: deploymentConfigService,
		releaseSetValidator:     releaseSetValidator,
		installationValidator:   installationValidator,
		deploymentValidator:     deploymentValidator,
		installerFactory:        installerFactory,
		releaseExtractor:        releaseExtractor,
		releaseManager:          releaseManager,
		releaseResolver:         releaseResolver,
		cloudFactory:            cloudFactory,
		agentClientFactory:      agentClientFactory,
		vmManagerFactory:        vmManagerFactory,
		stemcellExtractor:       stemcellExtractor,
		stemcellManagerFactory:  stemcellManagerFactory,
		deploymentRecord:        deploymentRecord,
		blobstoreFactory:        blobstoreFactory,
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

			var err error
			cpiRelease, err = c.releaseExtractor.Extract(releaseTarballPath)
			if err != nil {
				return bosherr.WrapErrorf(err, "Extracting release '%s'", releaseTarballPath)
			}
			c.releaseManager.Add(cpiRelease)
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

	var (
		deploymentManifest   bmdeplmanifest.Manifest
		installationManifest bminstallmanifest.Manifest
	)
	err = validationStage.PerformStep("Validating deployment manifest", func() error {
		releaseSetManifest, err := c.releaseSetParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
		}

		err = c.releaseSetValidator.Validate(releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating release set manifest")
		}

		//TODO: this seems to be a naming smell indicating a deeper issue
		c.releaseResolver.Filter(releaseSetManifest.Releases)

		deploymentManifest, err = c.deploymentParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", deploymentManifestPath)
		}

		err = c.deploymentValidator.Validate(deploymentManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		installationManifest, err = c.installationParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
		}

		err = c.installationValidator.Validate(installationManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating installation manifest")
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = validationStage.PerformStep("Validating cpi release", func() error {
		cpiRelease, err := c.releaseResolver.Find(installationManifest.Release)
		if err != nil {
			// should never happen, due to prior manifest validation
			return bosherr.WrapErrorf(err, "installation release '%s' must refer to a provided release", installationManifest.Release)
		}

		err = bmcpirel.NewValidator().Validate(cpiRelease)
		if err != nil {
			return bosherr.WrapErrorf(err, "Invalid CPI release '%s'", cpiRelease.Name())
		}

		return nil
	})
	if err != nil {
		return err
	}

	validationStage.Finish()

	isDeployed, err := c.deploymentRecord.IsDeployed(deploymentManifestPath, cpiRelease, extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed {
		c.ui.Sayln("No deployment, stemcell or cpi release changes. Skipping deploy.")
		return nil
	}

	installer, err := c.installerFactory.NewInstaller()
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI Installer")
	}

	installStage := c.eventLogger.NewStage("installing CPI")
	installStage.Start()

	installation, err := installer.Install(installationManifest, installStage)
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI")
	}

	err = installStage.PerformStep("Starting registry", func() error {
		return installation.StartRegistry()
	})
	if err != nil {
		return err
	}
	defer func() {
		err := installation.StopRegistry()
		if err != nil {
			c.logger.Warn(c.logTag, "Registry failed to stop: %s", err)
		}
	}()

	installStage.Finish()

	cloud, err := c.cloudFactory.NewCloud(installation, deploymentConfig.DirectorID)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	uploadStemcellStage := c.eventLogger.NewStage("uploading stemcell")
	uploadStemcellStage.Start()

	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell, uploadStemcellStage)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell")
	}

	uploadStemcellStage.Finish()

	agentClient := c.agentClientFactory.NewAgentClient(deploymentConfig.DirectorID, installationManifest.Mbus)
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient)

	blobstore, err := c.blobstoreFactory.Create(installationManifest.Mbus)
	if err != nil {
		return bosherr.WrapError(err, "Creating blobstore client")
	}

	deployStage := c.eventLogger.NewStage("deploying")
	deployStage.Start()

	_, err = c.deployer.Deploy(
		cloud,
		deploymentManifest,
		cloudStemcell,
		installationManifest.Registry,
		installationManifest.SSHTunnel,
		vmManager,
		blobstore,
		deployStage,
	)
	if err != nil {
		return bosherr.WrapError(err, "Deploying Microbosh")
	}

	err = c.deploymentRecord.Update(deploymentManifestPath, cpiRelease)
	if err != nil {
		return bosherr.WrapError(err, "Updating deployment record")
	}

	// TODO: cleanup unused disks here?

	if err = stemcellManager.DeleteUnused(deployStage); err != nil {
		return err
	}

	deployStage.Finish()

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

func (c *deployCmd) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
