package cmd

import (
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
	bitarball "github.com/cloudfoundry/bosh-init/installation/tarball"
	birel "github.com/cloudfoundry/bosh-init/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	biui "github.com/cloudfoundry/bosh-init/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

func NewDeploymentPreparer(
	ui biui.UI,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
	logTag string,
	deploymentStateService biconfig.DeploymentStateService,
	legacyDeploymentStateMigrator biconfig.LegacyDeploymentStateMigrator,
	releaseManager birel.Manager,
	deploymentRecord bidepl.Record,
	installerFactory biinstall.InstallerFactory,
	cloudFactory bicloud.Factory,
	stemcellManagerFactory bistemcell.ManagerFactory,
	agentClientFactory bihttpagent.AgentClientFactory,
	vmManagerFactory bivm.ManagerFactory,
	blobstoreFactory biblobstore.Factory,
	deployer bidepl.Deployer,
	releaseSetParser birelsetmanifest.Parser,
	installationParser biinstallmanifest.Parser,
	deploymentParser bideplmanifest.Parser,
	deploymentValidator bideplmanifest.Validator,
	releaseExtractor birel.Extractor,
	stemcellExtractor bistemcell.Extractor,
	deploymentManifestPath string,
	tarballProvider bitarball.Provider,

) DeploymentPreparer {
	return DeploymentPreparer{
		ui:                            ui,
		fs:                            fs,
		logger:                        logger,
		logTag:                        logTag,
		deploymentStateService:        deploymentStateService,
		legacyDeploymentStateMigrator: legacyDeploymentStateMigrator,
		releaseManager:                releaseManager,
		deploymentRecord:              deploymentRecord,
		installerFactory:              installerFactory,
		cloudFactory:                  cloudFactory,
		stemcellManagerFactory:        stemcellManagerFactory,
		agentClientFactory:            agentClientFactory,
		vmManagerFactory:              vmManagerFactory,
		blobstoreFactory:              blobstoreFactory,
		deployer:                      deployer,
		releaseSetParser:              releaseSetParser,
		installationParser:            installationParser,
		deploymentParser:              deploymentParser,
		deploymentValidator:           deploymentValidator,
		releaseExtractor:              releaseExtractor,
		stemcellExtractor:             stemcellExtractor,
		deploymentManifestPath:        deploymentManifestPath,
		tarballProvider:               tarballProvider,
	}
}

type DeploymentPreparer struct {
	ui                            biui.UI
	fs                            boshsys.FileSystem
	logger                        boshlog.Logger
	logTag                        string
	deploymentStateService        biconfig.DeploymentStateService
	legacyDeploymentStateMigrator biconfig.LegacyDeploymentStateMigrator
	releaseManager                birel.Manager
	deploymentRecord              bidepl.Record
	installerFactory              biinstall.InstallerFactory
	cloudFactory                  bicloud.Factory
	stemcellManagerFactory        bistemcell.ManagerFactory
	agentClientFactory            bihttpagent.AgentClientFactory
	vmManagerFactory              bivm.ManagerFactory
	blobstoreFactory              biblobstore.Factory
	deployer                      bidepl.Deployer
	releaseSetParser              birelsetmanifest.Parser
	installationParser            biinstallmanifest.Parser
	deploymentParser              bideplmanifest.Parser
	deploymentValidator           bideplmanifest.Validator
	releaseExtractor              birel.Extractor
	stemcellExtractor             bistemcell.Extractor
	deploymentManifestPath        string
	tarballProvider               bitarball.Provider
}

func (c *DeploymentPreparer) PrepareDeployment(stage biui.Stage) (err error) {
	c.ui.PrintLinef("Deployment state: '%s'", c.deploymentStateService.Path())

	if !c.deploymentStateService.Exists() {
		migrated, err := c.legacyDeploymentStateMigrator.MigrateIfExists(biconfig.LegacyDeploymentStatePath(c.deploymentManifestPath))
		if err != nil {
			return bosherr.WrapError(err, "Migrating legacy deployment state file")
		}
		if migrated {
			c.ui.PrintLinef("Migrated legacy deployments file: '%s'", biconfig.LegacyDeploymentStatePath(c.deploymentManifestPath))
		}
	}

	deploymentState, err := c.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment state")
	}

	defer func() {
		err := c.releaseManager.DeleteAll()
		if err != nil {
			c.logger.Warn(c.logTag, "Deleting all extracted releases: %s", err.Error())
		}
	}()

	var (
		extractedStemcell    bistemcell.ExtractedStemcell
		deploymentManifest   bideplmanifest.Manifest
		installationManifest biinstallmanifest.Manifest
	)

	installer, err := c.installerFactory.NewInstaller()
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI Installer")
	}

	cpiInstaller := bicpirel.CpiInstaller{
		ReleaseManager: c.releaseManager,
		Installer:      installer,
		Validator:      bicpirel.NewValidator(),
	}

	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		var releaseSetManifest birelsetmanifest.Manifest
		releaseSetManifest, installationManifest, err = releaseSetAndInstallationManifestParser{
			releaseSetParser:   c.releaseSetParser,
			installationParser: c.installationParser,
		}.ReleaseSetAndInstallationManifest(c.deploymentManifestPath)
		if err != nil {
			return err
		}

		for _, releaseRef := range releaseSetManifest.Releases {
			err = birel.NewFetcher(
				c.tarballProvider,
				c.releaseExtractor,
				c.releaseManager,
			).DownloadAndExtract(releaseRef, stage)
			if err != nil {
				return err
			}
		}

		err := cpiInstaller.ValidateCpiRelease(installationManifest, stage)
		if err != nil {
			return err
		}

		deploymentManifest, err = deploymentParser{
			deploymentParser:    c.deploymentParser,
			deploymentValidator: c.deploymentValidator,
			releaseManager:      c.releaseManager,
		}.GetDeploymentManifest(c.deploymentManifestPath, releaseSetManifest, stage)
		if err != nil {
			return err
		}

		extractedStemcell, err = bistemcell.Fetcher{
			TarballProvider:   c.tarballProvider,
			StemcellExtractor: c.stemcellExtractor,
		}.GetStemcell(deploymentManifest, stage)
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

	isDeployed, err := c.deploymentRecord.IsDeployed(c.deploymentManifestPath, c.releaseManager.List(), extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed {
		c.ui.PrintLinef("No deployment, stemcell or release changes. Skipping deploy.")
		return nil
	}

	installation, err := cpiInstaller.InstallCpiRelease(installationManifest, stage)
	if err != nil {
		return err
	}

	return installation.WithRunningRegistry(c.logger, stage, func() error {
		cloud, err := c.cloudFactory.NewCloud(installation, deploymentState.DirectorID)
		if err != nil {
			return bosherr.WrapError(err, "Creating CPI client from CPI installation")
		}

		stemcellManager := c.stemcellManagerFactory.NewManager(cloud)

		cloudStemcell, err := stemcellManager.Upload(extractedStemcell, stage)
		if err != nil {
			return err
		}

		agentClient := c.agentClientFactory.NewAgentClient(deploymentState.DirectorID, installationManifest.Mbus)
		vmManager := c.vmManagerFactory.NewManager(cloud, agentClient)

		blobstore, err := c.blobstoreFactory.Create(installationManifest.Mbus)
		if err != nil {
			return bosherr.WrapError(err, "Creating blobstore client")
		}

		err = stage.PerformComplex("deploying", func(deployStage biui.Stage) error {
			err = c.deploymentRecord.Clear()
			if err != nil {
				return bosherr.WrapError(err, "Clearing deployment record")
			}

			_, err = c.deployer.Deploy(
				cloud,
				deploymentManifest,
				cloudStemcell,
				installationManifest.Registry,
				vmManager,
				blobstore,
				deployStage,
			)
			if err != nil {
				return bosherr.WrapError(err, "Deploying")
			}

			err = c.deploymentRecord.Update(c.deploymentManifestPath, c.releaseManager.List())
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
	})
}
