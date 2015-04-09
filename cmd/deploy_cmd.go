package cmd

import (
	"errors"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-init/cloud"
	bmconfig "github.com/cloudfoundry/bosh-init/config"
	bmcpirel "github.com/cloudfoundry/bosh-init/cpi/release"
	bmdepl "github.com/cloudfoundry/bosh-init/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"
	bmdeplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bmvm "github.com/cloudfoundry/bosh-init/deployment/vm"
	bminstall "github.com/cloudfoundry/bosh-init/installation"
	bminstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-init/release"
	bmrelset "github.com/cloudfoundry/bosh-init/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-init/stemcell"
	bmui "github.com/cloudfoundry/bosh-init/ui"
)

type deployCmd struct {
	ui                             bmui.UI
	userConfig                     bmconfig.UserConfig
	fs                             boshsys.FileSystem
	releaseSetParser               bmrelsetmanifest.Parser
	installationParser             bminstallmanifest.Parser
	deploymentParser               bmdeplmanifest.Parser
	legacyDeploymentConfigMigrator bmconfig.LegacyDeploymentConfigMigrator
	deploymentConfigService        bmconfig.DeploymentConfigService
	releaseSetValidator            bmrelsetmanifest.Validator
	installationValidator          bminstallmanifest.Validator
	deploymentValidator            bmdeplmanifest.Validator
	installerFactory               bminstall.InstallerFactory
	releaseExtractor               bmrel.Extractor
	releaseManager                 bmrel.Manager
	releaseResolver                bmrelset.Resolver
	cloudFactory                   bmcloud.Factory
	agentClientFactory             bmhttpagent.AgentClientFactory
	vmManagerFactory               bmvm.ManagerFactory
	stemcellExtractor              bmstemcell.Extractor
	stemcellManagerFactory         bmstemcell.ManagerFactory
	deploymentRecord               bmdepl.Record
	blobstoreFactory               bmblobstore.Factory
	deployer                       bmdepl.Deployer
	eventLogger                    bmui.Stage
	logger                         boshlog.Logger
	logTag                         string
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	releaseSetParser bmrelsetmanifest.Parser,
	installationParser bminstallmanifest.Parser,
	deploymentParser bmdeplmanifest.Parser,
	legacyDeploymentConfigMigrator bmconfig.LegacyDeploymentConfigMigrator,
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
		logger:                         logger,
		logTag:                         "deployCmd",
	}
}

func (c *deployCmd) Name() string {
	return "deploy"
}

func (c *deployCmd) Run(stage bmui.Stage, args []string) error {
	stemcellTarballPath, releaseTarballPaths, err := c.parseCmdInputs(args)
	if err != nil {
		return err
	}

	deploymentManifestPath, err := getDeploymentManifest(c.userConfig, c.ui, c.fs)
	if err != nil {
		return err
	}

	if !c.deploymentConfigService.Exists() {
		migrated, err := c.legacyDeploymentConfigMigrator.MigrateIfExists()
		if err != nil {
			return bosherr.WrapError(err, "Migrating legacy deployment config file")
		}
		if migrated {
			c.ui.PrintLinef("Migrated legacy deployments file: '%s'", c.legacyDeploymentConfigMigrator.Path())
		}
	}

	deploymentConfig, err := c.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}

	var (
		extractedStemcell    bmstemcell.ExtractedStemcell
		deploymentManifest   bmdeplmanifest.Manifest
		installationManifest bminstallmanifest.Manifest
	)
	err = stage.PerformComplex("validating", func(stage bmui.Stage) error {
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

	var installation bminstall.Installation
	err = stage.PerformComplex("installing CPI", func(installStage bmui.Stage) error {
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

	err = stage.PerformComplex("deploying", func(deployStage bmui.Stage) error {
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

func (c *deployCmd) parseCmdInputs(args []string) (string, []string, error) {
	if len(args) < 2 {
		c.ui.ErrorLinef("Invalid usage - deploy command requires at least 2 arguments")
		c.ui.PrintLinef("Expected usage: bosh-init deploy <stemcell-tarball> <cpi-release-tarball> [release-2-tarball [release-3-tarball...]]")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", []string{}, errors.New("Invalid usage - deploy command requires at least 2 arguments")
	}
	return args[0], args[1:], nil
}

func (c *deployCmd) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (c *deployCmd) validate(
	validationStage bmui.Stage,
	stemcellTarballPath string,
	releaseTarballPaths []string,
	deploymentManifestPath string,
) (
	extractedStemcell bmstemcell.ExtractedStemcell,
	deploymentManifest bmdeplmanifest.Manifest,
	installationManifest bminstallmanifest.Manifest,
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
		err = bmcpirel.NewValidator().Validate(cpiRelease, cpiReleaseJobName)
		if err != nil {
			return bosherr.WrapErrorf(err, "Invalid CPI release '%s'", cpiReleaseName)
		}

		return nil
	})

	return extractedStemcell, deploymentManifest, installationManifest, err
}
