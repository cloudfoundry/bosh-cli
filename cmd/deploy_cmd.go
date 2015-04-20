package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

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
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type deployCmd struct {
	deploymentPreparerProvider func(deploymentManifestPath string) DeploymentPreparer
	ui                         biui.UI
	fs                         boshsys.FileSystem
	eventLogger                biui.Stage
	logger                     boshlog.Logger
	logTag                     string
}

func NewDeployCmd(
	ui biui.UI,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
	deploymentPreparerProvider func(deploymentManifestPath string) DeploymentPreparer,
) Cmd {
	return &deployCmd{
		ui: ui,
		fs: fs,
		deploymentPreparerProvider: deploymentPreparerProvider,
		logger: logger,
		logTag: "deployCmd",
	}
}

func (c *deployCmd) Name() string {
	return "deploy"
}

func (c *deployCmd) Meta() Meta {
	return Meta{
		Synopsis: "Create or update a deployment",
		Usage:    "<deployment_manifest_path> <stemcell_path> <cpi_release_path> [<release_paths...>]",
		Env:      genericEnv,
	}
}

func (c *deployCmd) Run(stage biui.Stage, args []string) error {
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

	deploymentPreparer := c.deploymentPreparerProvider(manifestAbsFilePath)
	return deploymentPreparer.PrepareDeployment(stage)
}

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
	releaseSetValidator birelsetmanifest.Validator,
	installationValidator biinstallmanifest.Validator,
	deploymentValidator bideplmanifest.Validator,
	releaseExtractor birel.Extractor,
	stemcellExtractor bistemcell.Extractor,
	deploymentManifestPath string,

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
		releaseSetValidator:           releaseSetValidator,
		installationValidator:         installationValidator,
		deploymentValidator:           deploymentValidator,
		releaseExtractor:              releaseExtractor,
		stemcellExtractor:             stemcellExtractor,
		deploymentManifestPath:        deploymentManifestPath,
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
	releaseSetValidator           birelsetmanifest.Validator
	installationValidator         biinstallmanifest.Validator
	deploymentValidator           bideplmanifest.Validator
	releaseExtractor              birel.Extractor
	stemcellExtractor             bistemcell.Extractor
	deploymentManifestPath        string
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

	var (
		extractedStemcell    bistemcell.ExtractedStemcell
		deploymentManifest   bideplmanifest.Manifest
		installationManifest biinstallmanifest.Manifest
	)
	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		extractedStemcell, deploymentManifest, installationManifest, err = c.validate(stage, c.deploymentManifestPath)
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

	isDeployed, err := c.deploymentRecord.IsDeployed(c.deploymentManifestPath, c.releaseManager.List(), extractedStemcell)
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
			installationManifest.SSHTunnel,
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
}

type Deployment struct{}

func (c *deployCmd) parseCmdInputs(args []string) (string, error) {
	if len(args) != 1 {
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", errors.New("Invalid usage - deploy command requires exactly 1 argument")
	}
	return args[0], nil
}

func (c *deployCmd) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (c *DeploymentPreparer) validate(
	validationStage biui.Stage,
	deploymentManifestPath string,
) (
	extractedStemcell bistemcell.ExtractedStemcell,
	deploymentManifest bideplmanifest.Manifest,
	installationManifest biinstallmanifest.Manifest,
	err error,
) {
	var releaseSetManifest birelsetmanifest.Manifest
	err = validationStage.Perform("Validating releases", func() error {
		releaseSetManifest, err = c.releaseSetParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
		}

		err = c.releaseSetValidator.Validate(releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating release set manifest")
		}

		for _, releaseRef := range releaseSetManifest.Releases {
			if !c.fs.FileExists(releaseRef.Path()) {
				return bosherr.Errorf("Verifying that the release '%s' exists", releaseRef.Path())
			}

			release, err := c.releaseExtractor.Extract(releaseRef.Path())
			if err != nil {
				return bosherr.WrapErrorf(err, "Extracting release '%s'", releaseRef.Path())
			}

			if release.Name() != releaseRef.Name {
				return bosherr.Errorf("Release name '%s' does not match the name in release tarball '%s'", releaseRef.Name, release.Name())
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

		err = c.installationValidator.Validate(installationManifest, releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating installation manifest")
		}

		return nil
	})
	if err != nil {
		return extractedStemcell, deploymentManifest, installationManifest, err
	}

	err = validationStage.Perform("Validating stemcell", func() error {
		stemcellTarballPath := deploymentManifest.ResourcePool().Stemcell.Path()
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

	err = validationStage.Perform("Validating cpi release", func() error {
		cpiReleaseName := installationManifest.Template.Release
		cpiRelease, found := c.releaseManager.Find(cpiReleaseName)
		if !found {
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
