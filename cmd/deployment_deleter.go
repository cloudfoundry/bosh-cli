package cmd

import (
	biblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bicpirel "github.com/cloudfoundry/bosh-init/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-init/deployment"
	bihttpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	birel "github.com/cloudfoundry/bosh-init/release"
	biui "github.com/cloudfoundry/bosh-init/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type DeploymentDeleter interface {
	DeleteDeployment(stage biui.Stage) (err error)
}

func NewDeploymentDeleter(
	ui biui.UI,
	logTag string,
	logger boshlog.Logger,
	fs boshsys.FileSystem,
	deploymentStateService biconfig.DeploymentStateService,
	releaseManager birel.Manager,
	installerFactory biinstall.InstallerFactory,
	cloudFactory bicloud.Factory,
	agentClientFactory bihttpagent.AgentClientFactory,
	blobstoreFactory biblobstore.Factory,
	deploymentManagerFactory bidepl.ManagerFactory,
	installationParser biinstallmanifest.Parser,
	deploymentManifestPath string,
	cpiReleaseValidator bicpirel.CPIReleaseValidator,
	validatedCpiReleaseSpec bicpirel.ValidatedCpiReleaseSpec,
) DeploymentDeleter {
	return &deploymentDeleter{
		ui:     ui,
		logTag: logTag,
		logger: logger,
		fs:     fs,
		deploymentStateService:   deploymentStateService,
		releaseManager:           releaseManager,
		installerFactory:         installerFactory,
		cloudFactory:             cloudFactory,
		agentClientFactory:       agentClientFactory,
		blobstoreFactory:         blobstoreFactory,
		deploymentManagerFactory: deploymentManagerFactory,
		installationParser:       installationParser,
		deploymentManifestPath:   deploymentManifestPath,
		cpiReleaseValidator:      cpiReleaseValidator,
		validatedCpiReleaseSpec:  validatedCpiReleaseSpec,
	}
}

type deploymentDeleter struct {
	ui                       biui.UI
	logTag                   string
	logger                   boshlog.Logger
	fs                       boshsys.FileSystem
	deploymentStateService   biconfig.DeploymentStateService
	releaseManager           birel.Manager
	installerFactory         biinstall.InstallerFactory
	cloudFactory             bicloud.Factory
	agentClientFactory       bihttpagent.AgentClientFactory
	blobstoreFactory         biblobstore.Factory
	deploymentManagerFactory bidepl.ManagerFactory
	installationParser       biinstallmanifest.Parser
	deploymentManifestPath   string
	cpiReleaseValidator      bicpirel.CPIReleaseValidator
	validatedCpiReleaseSpec  bicpirel.ValidatedCpiReleaseSpec
}

func (c *deploymentDeleter) DeleteDeployment(stage biui.Stage) (err error) {
	c.ui.PrintLinef("Deployment state: '%s'", c.deploymentStateService.Path())

	if !c.deploymentStateService.Exists() {
		c.ui.PrintLinef("No deployment state file found.")
		return nil
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

	installationManifest, installation, err := c.installCPI(stage)
	if err != nil {
		return err
	}

	return installation.WithRunningRegistry(c.logger, stage, func() error {
		return c.findAndDeleteDeployment(stage, installation, deploymentState.DirectorID, installationManifest.Mbus)
	})
}

func (c *deploymentDeleter) findAndDeleteDeployment(stage biui.Stage, installation biinstall.Installation, directorID, installationMbus string) error {
	deploymentManager, err := c.deploymentManager(installation, directorID, installationMbus)
	if err != nil {
		return err
	}
	err = c.findCurrentDeploymentAndDelete(stage, deploymentManager)
	if err != nil {
		return bosherr.WrapError(err, "Deleting deployment")
	}
	return deploymentManager.Cleanup(stage)
}

func (c *deploymentDeleter) findCurrentDeploymentAndDelete(stage biui.Stage, deploymentManager bidepl.Manager) error {
	c.logger.Debug(c.logTag, "Finding current deployment...")
	deployment, found, err := deploymentManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment")
	}

	return stage.PerformComplex("deleting deployment", func(deleteStage biui.Stage) error {
		if !found {
			//TODO: skip? would require adding skip support to PerformComplex
			c.logger.Debug(c.logTag, "No current deployment found...")
			return nil
		}

		return deployment.Delete(deleteStage)
	})
}

func (c *deploymentDeleter) installCPI(stage biui.Stage) (biinstallmanifest.Manifest, biinstall.Installation, error) {
	installationManifest, err := c.getInstallationManifestAndRegisterValidCpiRelease(stage)
	if err != nil {
		return installationManifest, nil, err
	}

	installer, err := c.installerFactory.NewInstaller()
	if err != nil {
		return installationManifest, nil, bosherr.WrapError(err, "Creating CPI Installer")
	}

	var installation biinstall.Installation
	err = stage.PerformComplex("installing CPI", func(installStage biui.Stage) error {
		installation, err = installer.Install(installationManifest, installStage)
		return err
	})
	return installationManifest, installation, err
}

func (c *deploymentDeleter) getInstallationManifestAndRegisterValidCpiRelease(stage biui.Stage) (biinstallmanifest.Manifest, error) {
	var installationManifest biinstallmanifest.Manifest
	err := stage.PerformComplex("validating", func(stage biui.Stage) error {
		var err error
		installationManifest, err = c.getInstallationManifestFrom(c.deploymentManifestPath, stage)
		if err != nil {
			return err
		}

		cpiReleaseRef, err := c.validatedCpiReleaseSpec.GetFrom(c.deploymentManifestPath, installationManifest)
		if err != nil {
			return err
		}

		return c.cpiReleaseValidator.DownloadAndRegister(cpiReleaseRef, installationManifest, stage)
	})
	return installationManifest, err
}

func (c *deploymentDeleter) getInstallationManifestFrom(deploymentManifestPath string, stage biui.Stage) (biinstallmanifest.Manifest, error) {
	var installationManifest biinstallmanifest.Manifest
	var err error
	err = stage.Perform("Validating deployment manifest", func() error {
		installationManifest, err = c.installationParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
		}
		return err
	})
	return installationManifest, err
}

func (c *deploymentDeleter) deploymentManager(installation biinstall.Installation, directorID, installationMbus string) (bidepl.Manager, error) {
	c.logger.Debug(c.logTag, "Creating cloud client...")
	cloud, err := c.cloudFactory.NewCloud(installation, directorID)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	c.logger.Debug(c.logTag, "Creating agent client...")
	agentClient := c.agentClientFactory.NewAgentClient(directorID, installationMbus)

	c.logger.Debug(c.logTag, "Creating blobstore client...")
	blobstore, err := c.blobstoreFactory.Create(installationMbus)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating blobstore client")
	}

	c.logger.Debug(c.logTag, "Creating deployment manager...")
	return c.deploymentManagerFactory.NewManager(cloud, agentClient, blobstore), nil
}
