package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	"github.com/cloudfoundry/bosh-agent/uuid"
	biblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bicpirel "github.com/cloudfoundry/bosh-init/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-init/deployment"
	bihttpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bivm "github.com/cloudfoundry/bosh-init/deployment/vm"
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	birel "github.com/cloudfoundry/bosh-init/release"
	birelset "github.com/cloudfoundry/bosh-init/release/set"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type deployCmd struct {
	ui                             biui.UI
	userConfig                     biconfig.UserConfig
	fs                             boshsys.FileSystem
	releaseSetParser               birelsetmanifest.Parser
	installationParser             biinstallmanifest.Parser
	deploymentParser               bideplmanifest.Parser
	legacyDeploymentConfigMigrator biconfig.LegacyDeploymentConfigMigrator
	deploymentConfigService        biconfig.DeploymentConfigService
	releaseSetValidator            birelsetmanifest.Validator
	installationValidator          biinstallmanifest.Validator
	deploymentValidator            bideplmanifest.Validator
	installerFactory               biinstall.InstallerFactory
	releaseExtractor               birel.Extractor
	releaseManager                 birel.Manager
	releaseResolver                birelset.Resolver
	cloudFactory                   bicloud.Factory
	agentClientFactory             bihttpagent.AgentClientFactory
	vmManagerFactory               bivm.ManagerFactory
	stemcellExtractor              bistemcell.Extractor
	stemcellManagerFactory         bistemcell.ManagerFactory
	deploymentRecord               bidepl.Record
	blobstoreFactory               biblobstore.Factory
	deployer                       bidepl.Deployer
	eventLogger                    biui.Stage
	uuidGenerator                  uuid.Generator
	logger                         boshlog.Logger
	logTag                         string
}

func NewDeployCmd(
	ui biui.UI,
	userConfig biconfig.UserConfig,
	fs boshsys.FileSystem,
	releaseSetParser birelsetmanifest.Parser,
	installationParser biinstallmanifest.Parser,
	deploymentParser bideplmanifest.Parser,
	legacyDeploymentConfigMigrator biconfig.LegacyDeploymentConfigMigrator,
	deploymentConfigService biconfig.DeploymentConfigService,
	releaseSetValidator birelsetmanifest.Validator,
	installationValidator biinstallmanifest.Validator,
	deploymentValidator bideplmanifest.Validator,
	installerFactory biinstall.InstallerFactory,
	releaseExtractor birel.Extractor,
	releaseManager birel.Manager,
	releaseResolver birelset.Resolver,
	cloudFactory bicloud.Factory,
	agentClientFactory bihttpagent.AgentClientFactory,
	vmManagerFactory bivm.ManagerFactory,
	stemcellExtractor bistemcell.Extractor,
	stemcellManagerFactory bistemcell.ManagerFactory,
	deploymentRecord bidepl.Record,
	blobstoreFactory biblobstore.Factory,
	deployer bidepl.Deployer,
	uuidGenerator uuid.Generator,
	logger boshlog.Logger,
) Cmd {
	return &deployCmd{
		ui:                             ui,
		userConfig:                     userConfig,
		fs:                             fs,
		releaseSetParser:               releaseSetParser,
		installationParser:             installationParser,
		deploymentParser:               deploymentParser,
		legacyDeploymentConfigMigrator: legacyDeploymentConfigMigrator,
		deploymentConfigService:        deploymentConfigService,
		releaseSetValidator:            releaseSetValidator,
		installationValidator:          installationValidator,
		deploymentValidator:            deploymentValidator,
		installerFactory:               installerFactory,
		releaseExtractor:               releaseExtractor,
		releaseManager:                 releaseManager,
		releaseResolver:                releaseResolver,
		cloudFactory:                   cloudFactory,
		agentClientFactory:             agentClientFactory,
		vmManagerFactory:               vmManagerFactory,
		stemcellExtractor:              stemcellExtractor,
		stemcellManagerFactory:         stemcellManagerFactory,
		deploymentRecord:               deploymentRecord,
		blobstoreFactory:               blobstoreFactory,
		deployer:                       deployer,
		uuidGenerator:                  uuidGenerator,
		logger:                         logger,
		logTag:                         "deployCmd",
	}
}

func (c *deployCmd) Name() string {
	return "deploy"
}

func (c *deployCmd) Run(stage biui.Stage, args []string) error {
	deploymentManifestPath, stemcellTarballPath, releaseTarballPaths, err := c.parseCmdInputs(args)
	if err != nil {
		return err
	}

	manifestAbsFilePath, err := filepath.Abs(deploymentManifestPath)
	if err != nil {
		c.ui.ErrorLinef("Failed getting absolute path to deployment file '%s'", deploymentManifestPath)
		return bosherr.WrapErrorf(err, "Getting absolute path to deployment file '%s'", deploymentManifestPath)
	}

	if !c.fs.FileExists(manifestAbsFilePath) {
		c.ui.ErrorLinef("Deployment '%s' does not exist", manifestAbsFilePath)
		return bosherr.Errorf("Deployment manifest does not exist at '%s'", manifestAbsFilePath)
	}

	c.userConfig.DeploymentManifestPath = manifestAbsFilePath
	c.ui.PrintLinef("Deployment manifest: '%s'", manifestAbsFilePath)

	deploymentConfigPath := c.userConfig.DeploymentConfigPath()
	c.deploymentConfigService.SetConfigPath(deploymentConfigPath)

	c.ui.PrintLinef("Deployment state: '%s'", deploymentConfigPath)

	if !c.deploymentConfigService.Exists() {
		migrated, err := c.legacyDeploymentConfigMigrator.MigrateIfExists(c.userConfig.LegacyDeploymentConfigPath())
		if err != nil {
			return bosherr.WrapError(err, "Migrating legacy deployment config file")
		}
		if migrated {
			c.ui.PrintLinef("Migrated legacy deployments file: '%s'", c.userConfig.LegacyDeploymentConfigPath())
		}
	}

	deploymentConfig, err := c.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}

	var (
		extractedStemcell    bistemcell.ExtractedStemcell
		deploymentManifest   bideplmanifest.Manifest
		installationManifest biinstallmanifest.Manifest
	)
	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		extractedStemcell, deploymentManifest, installationManifest, err = c.validate(stage, stemcellTarballPath, releaseTarballPaths, deploymentManifestPath)
		return err
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
	defer func() {
		err := c.releaseManager.DeleteAll()
		if err != nil {
			c.logger.Warn(c.logTag, "Deleting all extracted releases: %s", err.Error())
		}
	}()

	isDeployed, err := c.deploymentRecord.IsDeployed(deploymentManifestPath, c.releaseManager.List(), extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed {
		c.ui.PrintLinef("No deployment, stemcell or release changes. Skipping deploy.")
		return nil
	}

	installer, err := c.installerFactory.NewInstaller()
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI Installer")
	}

	var installation biinstall.Installation
	err = stage.PerformComplex("installing CPI", func(installStage biui.Stage) error {
		installation, err = installer.Install(installationManifest, installStage)
		return err
	})
	if err != nil {
		return err
	}

	err = stage.Perform("Starting registry", func() error {
		return installation.StartRegistry()
	})
	if err != nil {
		return err
	}
	defer func() {
		//TODO: wrap stopping registry in stage?
		err := installation.StopRegistry()
		if err != nil {
			c.logger.Warn(c.logTag, "Registry failed to stop: %s", err)
		}
	}()

	cloud, err := c.cloudFactory.NewCloud(installation, deploymentConfig.DirectorID)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)

	cloudStemcell, err := stemcellManager.Upload(extractedStemcell, stage)
	if err != nil {
		return err
	}

	agentClient := c.agentClientFactory.NewAgentClient(deploymentConfig.DirectorID, installationManifest.Mbus)
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient)

	blobstore, err := c.blobstoreFactory.Create(installationManifest.Mbus)
	if err != nil {
		return bosherr.WrapError(err, "Creating blobstore client")
	}

	err = stage.PerformComplex("deploying", func(deployStage biui.Stage) error {
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
			return bosherr.WrapError(err, "Deploying")
		}

		err = c.deploymentRecord.Update(deploymentManifestPath, c.releaseManager.List())
		if err != nil {
			return bosherr.WrapError(err, "Updating deployment record")
		}

		return nil
	})
	if err != nil {
		return err
	}

	// TODO: cleanup unused disks here?

	err = stemcellManager.DeleteUnused(stage)
	if err != nil {
		return err
	}

	return nil
}

type Deployment struct{}

func (c *deployCmd) parseCmdInputs(args []string) (string, string, []string, error) {
	if len(args) < 3 {
		c.ui.ErrorLinef("Invalid usage - deploy command requires at least 3 arguments")
		c.ui.PrintLinef("Expected usage: bosh-init deploy <deployment-manifest> <stemcell-tarball> <cpi-release-tarball> [release-2-tarball [release-3-tarball...]]")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", "", []string{}, errors.New("Invalid usage - deploy command requires at least 3 arguments")
	}
	return args[0], args[1], args[2:], nil
}

func (c *deployCmd) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (c *deployCmd) validate(
	validationStage biui.Stage,
	stemcellTarballPath string,
	releaseTarballPaths []string,
	deploymentManifestPath string,
) (
	extractedStemcell bistemcell.ExtractedStemcell,
	deploymentManifest bideplmanifest.Manifest,
	installationManifest biinstallmanifest.Manifest,
	err error,
) {
	err = validationStage.Perform("Validating stemcell", func() error {
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
		return extractedStemcell, deploymentManifest, installationManifest, err
	}
	defer func() {
		if err != nil {
			deleteErr := extractedStemcell.Delete()
			if deleteErr != nil {
				c.logger.Warn(c.logTag, "Failed to delete extracted stemcell: %s", deleteErr.Error())
			}
		}
	}()

	err = validationStage.Perform("Validating releases", func() error {
		for _, releaseTarballPath := range releaseTarballPaths {
			if !c.fs.FileExists(releaseTarballPath) {
				return bosherr.Errorf("Verifying that the release '%s' exists", releaseTarballPath)
			}

			release, err := c.releaseExtractor.Extract(releaseTarballPath)
			if err != nil {
				return bosherr.WrapErrorf(err, "Extracting release '%s'", releaseTarballPath)
			}
			c.releaseManager.Add(release)
		}

		return nil
	})
	if err != nil {
		return extractedStemcell, deploymentManifest, installationManifest, err
	}
	defer func() {
		if err != nil {
			err := c.releaseManager.DeleteAll()
			if err != nil {
				c.logger.Warn(c.logTag, "Deleting all extracted releases: %s", err.Error())
			}
		}
	}()

	err = validationStage.Perform("Validating deployment manifest", func() error {
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
		return extractedStemcell, deploymentManifest, installationManifest, err
	}

	err = validationStage.Perform("Validating cpi release", func() error {
		cpiReleaseName := installationManifest.Template.Release
		cpiRelease, err := c.releaseResolver.Find(cpiReleaseName)
		if err != nil {
			// should never happen, due to prior manifest validation
			return bosherr.WrapErrorf(err, "installation release '%s' must refer to a provided release", cpiReleaseName)
		}

		cpiReleaseJobName := installationManifest.Template.Name
		err = bicpirel.NewValidator().Validate(cpiRelease, cpiReleaseJobName)
		if err != nil {
			return bosherr.WrapErrorf(err, "Invalid CPI release '%s'", cpiReleaseName)
		}

		return nil
	})

	return extractedStemcell, deploymentManifest, installationManifest, err
}
