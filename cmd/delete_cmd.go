package cmd

import (
	"errors"
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
	birel "github.com/cloudfoundry/bosh-init/release"
	birelset "github.com/cloudfoundry/bosh-init/release/set"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type deleteCmd struct {
	ui                       biui.UI
	fs                       boshsys.FileSystem
	releaseSetParser         birelsetmanifest.Parser
	installationParser       biinstallmanifest.Parser
	deploymentConfigService  biconfig.DeploymentConfigService
	releaseSetValidator      birelsetmanifest.Validator
	installationValidator    biinstallmanifest.Validator
	installerFactory         biinstall.InstallerFactory
	releaseExtractor         birel.Extractor
	releaseManager           birel.Manager
	releaseResolver          birelset.Resolver
	cloudFactory             bicloud.Factory
	agentClientFactory       bihttpagent.AgentClientFactory
	blobstoreFactory         biblobstore.Factory
	deploymentManagerFactory bidepl.ManagerFactory
	logger                   boshlog.Logger
	logTag                   string
}

func NewDeleteCmd(
	ui biui.UI,
	fs boshsys.FileSystem,
	releaseSetParser birelsetmanifest.Parser,
	installationParser biinstallmanifest.Parser,
	deploymentConfigService biconfig.DeploymentConfigService,
	releaseSetValidator birelsetmanifest.Validator,
	installationValidator biinstallmanifest.Validator,
	installerFactory biinstall.InstallerFactory,
	releaseExtractor birel.Extractor,
	releaseManager birel.Manager,
	releaseResolver birelset.Resolver,
	cloudFactory bicloud.Factory,
	agentClientFactory bihttpagent.AgentClientFactory,
	blobstoreFactory biblobstore.Factory,
	deploymentManagerFactory bidepl.ManagerFactory,
	logger boshlog.Logger,
) Cmd {
	return &deleteCmd{
		ui:                       ui,
		fs:                       fs,
		releaseSetParser:         releaseSetParser,
		installationParser:       installationParser,
		deploymentConfigService:  deploymentConfigService,
		releaseSetValidator:      releaseSetValidator,
		installationValidator:    installationValidator,
		installerFactory:         installerFactory,
		releaseExtractor:         releaseExtractor,
		releaseManager:           releaseManager,
		releaseResolver:          releaseResolver,
		cloudFactory:             cloudFactory,
		agentClientFactory:       agentClientFactory,
		blobstoreFactory:         blobstoreFactory,
		deploymentManagerFactory: deploymentManagerFactory,
		logger: logger,
		logTag: "deleteCmd",
	}
}

func (c *deleteCmd) Name() string {
	return "delete"
}

func (c *deleteCmd) Run(stage biui.Stage, args []string) error {
	deploymentManifestPath, releaseTarballPath, err := c.parseCmdInputs(args)
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

	deploymentConfigPath := biconfig.DeploymentConfigPath(manifestAbsFilePath)
	c.deploymentConfigService.SetConfigPath(deploymentConfigPath)

	c.ui.PrintLinef("Deployment state: '%s'", deploymentConfigPath)

	if !c.deploymentConfigService.Exists() {
		c.ui.PrintLinef("No deployment config file found.")
		return nil
	}

	deploymentConfig, err := c.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}

	var installationManifest biinstallmanifest.Manifest
	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		installationManifest, err = c.validate(stage, releaseTarballPath, deploymentManifestPath)
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
	cloud, err := c.cloudFactory.NewCloud(installation, deploymentConfig.DirectorID)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	c.logger.Debug(c.logTag, "Creating agent client...")
	agentClient := c.agentClientFactory.NewAgentClient(deploymentConfig.DirectorID, installationManifest.Mbus)

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

func (c *deleteCmd) parseCmdInputs(args []string) (string, string, error) {
	if len(args) != 2 {
		c.ui.ErrorLinef("Invalid usage - delete command requires exactly 2 arguments")
		c.ui.PrintLinef("Expected usage: bosh-init delete <deployment-manifest> <cpi-release-tarball>")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", "", errors.New("Invalid usage - delete command requires exactly 2 arguments")
	}
	return args[0], args[1], nil
}

func (c *deleteCmd) validate(validationStage biui.Stage, releaseTarballPath, deploymentManifestPath string) (
	installationManifest biinstallmanifest.Manifest,
	err error,
) {
	err = validationStage.Perform("Validating releases", func() error {
		if !c.fs.FileExists(releaseTarballPath) {
			return bosherr.Errorf("Verifying that the release '%s' exists", releaseTarballPath)
		}

		release, err := c.releaseExtractor.Extract(releaseTarballPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting release '%s'", releaseTarballPath)
		}
		c.releaseManager.Add(release)

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

	err = validationStage.Perform("Validating deployment manifest", func() error {
		releaseSetManifest, err := c.releaseSetParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
		}

		err = c.releaseSetValidator.Validate(releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating release set manifest")
		}

		c.releaseResolver.Filter(releaseSetManifest.Releases)

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
		return installationManifest, err
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

	return installationManifest, err
}
