package cmd

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpirel "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deleteCmd struct {
	ui                       bmui.UI
	userConfig               bmconfig.UserConfig
	fs                       boshsys.FileSystem
	releaseSetParser         bmrelsetmanifest.Parser
	installationParser       bminstallmanifest.Parser
	deploymentConfigService  bmconfig.DeploymentConfigService
	releaseSetValidator      bmrelsetmanifest.Validator
	installationValidator    bminstallmanifest.Validator
	installerFactory         bminstall.InstallerFactory
	releaseExtractor         bmrel.Extractor
	releaseManager           bmrel.Manager
	releaseResolver          bmrelset.Resolver
	cloudFactory             bmcloud.Factory
	agentClientFactory       bmhttpagent.AgentClientFactory
	blobstoreFactory         bmblobstore.Factory
	deploymentManagerFactory bmdepl.ManagerFactory
	logger                   boshlog.Logger
	logTag                   string
}

func NewDeleteCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	releaseSetParser bmrelsetmanifest.Parser,
	installationParser bminstallmanifest.Parser,
	deploymentConfigService bmconfig.DeploymentConfigService,
	releaseSetValidator bmrelsetmanifest.Validator,
	installationValidator bminstallmanifest.Validator,
	installerFactory bminstall.InstallerFactory,
	releaseExtractor bmrel.Extractor,
	releaseManager bmrel.Manager,
	releaseResolver bmrelset.Resolver,
	cloudFactory bmcloud.Factory,
	agentClientFactory bmhttpagent.AgentClientFactory,
	blobstoreFactory bmblobstore.Factory,
	deploymentManagerFactory bmdepl.ManagerFactory,
	logger boshlog.Logger,
) Cmd {
	return &deleteCmd{
		ui:                       ui,
		userConfig:               userConfig,
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

func (c *deleteCmd) Run(stage bmui.Stage, args []string) error {
	releaseTarballPath, err := c.parseCmdInputs(args)
	if err != nil {
		return err
	}

	deploymentManifestPath, err := getDeploymentManifest(c.userConfig, c.ui, c.fs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running delete cmd")
	}

	deploymentConfig, err := c.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}

	var installationManifest bminstallmanifest.Manifest
	err = stage.PerformComplex("validating", func(stage bmui.Stage) error {
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

	var installation bminstall.Installation
	err = stage.PerformComplex("installing CPI", func(installStage bmui.Stage) error {
		installation, err = installer.Install(installationManifest, installStage)
		return err
	})

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

	err = stage.PerformComplex("deleting deployment", func(deleteStage bmui.Stage) error {
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
		c.ui.ErrorLinef("Invalid usage - delete command requires exactly 1 argument")
		c.ui.PrintLinef("Expected usage: bosh-micro delete <cpi-release-tarball>")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", errors.New("Invalid usage - delete command requires exactly 1 argument")
	}
	return args[0], nil
}

func (c *deleteCmd) validate(validationStage bmui.Stage, releaseTarballPath, deploymentManifestPath string) (
	installationManifest bminstallmanifest.Manifest,
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

	return installationManifest, err
}
