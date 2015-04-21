package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	biblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bicpirel "github.com/cloudfoundry/bosh-init/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-init/deployment"
	bihttpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	bitarball "github.com/cloudfoundry/bosh-init/installation/tarball"
	birel "github.com/cloudfoundry/bosh-init/release"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type deleteCmd struct {
	deploymentDeleterProvider func(deploymentManifestString string) DeploymentDeleter
	ui                        biui.UI
	fs                        boshsys.FileSystem
	logger                    boshlog.Logger
	logTag                    string
}

func NewDeleteCmd(
	ui biui.UI,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
	deploymentDeleterProvider func(deploymentManifestString string) DeploymentDeleter,
) Cmd {
	return &deleteCmd{
		ui: ui,
		fs: fs,
		deploymentDeleterProvider: deploymentDeleterProvider,
		logger: logger,
		logTag: "deleteCmd",
	}
}

func (c *deleteCmd) Name() string {
	return "delete"
}

func (c *deleteCmd) Meta() Meta {
	return Meta{
		Synopsis: "Delete existing deployment",
		Usage:    "<deployment_manifest_path> <cpi_release_path>",
		Env:      genericEnv,
	}
}

func (c *deleteCmd) Run(stage biui.Stage, args []string) error {
	deploymentManifestPath, err := c.parseCmdInputs(args)
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

	c.ui.PrintLinef("Deployment manifest: '%s'", manifestAbsFilePath)

	deploymentDeleter := c.deploymentDeleterProvider(manifestAbsFilePath)
	return deploymentDeleter.DeleteDeployment(stage)
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
	releaseSetParser birelsetmanifest.Parser,
	releaseSetValidator birelsetmanifest.Validator,
	releaseExtractor birel.Extractor,
	installationParser biinstallmanifest.Parser,
	installationValidator biinstallmanifest.Validator,
	deploymentManifestPath string,
	tarballProvider bitarball.Provider,

) DeploymentDeleter {
	return DeploymentDeleter{
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
		releaseSetParser:         releaseSetParser,
		releaseSetValidator:      releaseSetValidator,
		releaseExtractor:         releaseExtractor,
		installationParser:       installationParser,
		installationValidator:    installationValidator,
		deploymentManifestPath:   deploymentManifestPath,
		tarballProvider:          tarballProvider,
	}
}

type DeploymentDeleter struct {
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
	releaseSetParser         birelsetmanifest.Parser
	releaseSetValidator      birelsetmanifest.Validator
	releaseExtractor         birel.Extractor
	installationParser       biinstallmanifest.Parser
	installationValidator    biinstallmanifest.Validator
	deploymentManifestPath   string
	tarballProvider          bitarball.Provider
}

func (c *DeploymentDeleter) DeleteDeployment(stage biui.Stage) (err error) {
	c.ui.PrintLinef("Deployment state: '%s'", c.deploymentStateService.Path())

	if !c.deploymentStateService.Exists() {
		c.ui.PrintLinef("No deployment state file found.")
		return nil
	}

	deploymentState, err := c.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment state")
	}

	var installationManifest biinstallmanifest.Manifest
	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		installationManifest, err = c.validate(stage, c.deploymentManifestPath)
		return err
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

	c.logger.Debug(c.logTag, "Creating cloud client...")
	cloud, err := c.cloudFactory.NewCloud(installation, deploymentState.DirectorID)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	c.logger.Debug(c.logTag, "Creating agent client...")
	agentClient := c.agentClientFactory.NewAgentClient(deploymentState.DirectorID, installationManifest.Mbus)

	c.logger.Debug(c.logTag, "Creating blobstore client...")
	blobstore, err := c.blobstoreFactory.Create(installationManifest.Mbus)
	if err != nil {
		return bosherr.WrapError(err, "Creating blobstore client")
	}

	c.logger.Debug(c.logTag, "Creating deployment manager...")
	deploymentManager := c.deploymentManagerFactory.NewManager(cloud, agentClient, blobstore)

	c.logger.Debug(c.logTag, "Finding current deployment...")
	deployment, found, err := deploymentManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment")
	}

	err = stage.PerformComplex("deleting deployment", func(deleteStage biui.Stage) error {
		if !found {
			//TODO: skip? would require adding skip support to PerformComplex
			c.logger.Debug(c.logTag, "No current deployment found...")
			return nil
		}

		err = deployment.Delete(deleteStage)
		if err != nil {
			return bosherr.WrapError(err, "Deleting deployment")
		}
		return nil
	})

	if err = deploymentManager.Cleanup(stage); err != nil {
		return err
	}

	return err
}

func (c *deleteCmd) parseCmdInputs(args []string) (string, error) {
	if len(args) != 1 {
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", errors.New("Invalid usage - delete command requires exactly 1 argument")
	}
	return args[0], nil
}

func (c *DeploymentDeleter) validate(validationStage biui.Stage, deploymentManifestPath string) (
	installationManifest biinstallmanifest.Manifest,
	err error,
) {
	var cpiRelease birelmanifest.ReleaseRef
	var releaseSetManifest birelsetmanifest.Manifest

	err = validationStage.Perform("Validating deployment manifest", func() error {
		installationManifest, err = c.installationParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
		}

		releaseSetManifest, err = c.releaseSetParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
		}

		err = c.releaseSetValidator.Validate(releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating release set manifest")
		}

		err = c.installationValidator.Validate(installationManifest, releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating installation manifest")
		}
		cpiReleaseName := installationManifest.Template.Release

		var found bool
		cpiRelease, found = releaseSetManifest.FindByName(cpiReleaseName)
		if !found {
			return bosherr.Errorf("installation release '%s' must refer to a release in releases", cpiReleaseName)
		}

		return nil
	})
	if err != nil {
		return installationManifest, err
	}

	releasePath, err := c.tarballProvider.Get(bitarball.Source(cpiRelease), validationStage)
	if err != nil {
		return installationManifest, err
	}

	err = validationStage.Perform(fmt.Sprintf("Validating release '%s'", cpiRelease.Name), func() error {
		release, err := c.releaseExtractor.Extract(releasePath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting release '%s'", releasePath)
		}
		c.releaseManager.Add(release)

		cpiReleaseJobName := installationManifest.Template.Name
		err = bicpirel.NewValidator().Validate(release, cpiReleaseJobName)
		if err != nil {
			return bosherr.WrapErrorf(err, "Invalid CPI release '%s'", release.Name())
		}

		return nil
	})
	if err != nil {
		return installationManifest, err
	}
	defer func() {
		if err != nil {
			err := c.releaseManager.DeleteAll()
			if err != nil {
				c.logger.Warn(c.logTag, "Deleting all extracted releases: %s", err.Error())
			}
		}
	}()

	return installationManifest, err
}
