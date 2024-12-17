package cmd

import (
	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	bihttpclient "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/go-patch/patch"

	biblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-cli/v7/deployment"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	boshinst "github.com/cloudfoundry/bosh-cli/v7/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

func NewDeploymentPreparer(
	ui biui.UI,
	logger boshlog.Logger,
	logTag string,
	deploymentStateService biconfig.DeploymentStateService,
	legacyDeploymentStateMigrator biconfig.LegacyDeploymentStateMigrator,
	releaseManager boshinst.ReleaseManager,
	deploymentRecord bidepl.Record,
	cloudFactory bicloud.Factory,
	stemcellManagerFactory bistemcell.ManagerFactory,
	agentClientFactory bihttpagent.AgentClientFactory,
	vmManagerFactory bivm.ManagerFactory,
	blobstoreFactory biblobstore.Factory,
	deployer bidepl.Deployer,
	deploymentManifestPath string,
	deploymentVars boshtpl.Variables,
	deploymentOp patch.Op,
	cpiInstaller bicpirel.CpiInstaller,
	releaseFetcher boshinst.ReleaseFetcher,
	stemcellFetcher bistemcell.Fetcher,
	releaseSetAndInstallationManifestParser ReleaseSetAndInstallationManifestParser,
	deploymentManifestParser DeploymentManifestParser,
	tempRootConfigurator TempRootConfigurator,
	targetProvider biinstall.TargetProvider,
) DeploymentPreparer {
	return DeploymentPreparer{
		ui:                                      ui,
		logger:                                  logger,
		logTag:                                  logTag,
		deploymentStateService:                  deploymentStateService,
		legacyDeploymentStateMigrator:           legacyDeploymentStateMigrator,
		releaseManager:                          releaseManager,
		deploymentRecord:                        deploymentRecord,
		cloudFactory:                            cloudFactory,
		stemcellManagerFactory:                  stemcellManagerFactory,
		agentClientFactory:                      agentClientFactory,
		vmManagerFactory:                        vmManagerFactory,
		blobstoreFactory:                        blobstoreFactory,
		deployer:                                deployer,
		deploymentManifestPath:                  deploymentManifestPath,
		deploymentVars:                          deploymentVars,
		deploymentOp:                            deploymentOp,
		cpiInstaller:                            cpiInstaller,
		releaseFetcher:                          releaseFetcher,
		stemcellFetcher:                         stemcellFetcher,
		releaseSetAndInstallationManifestParser: releaseSetAndInstallationManifestParser,
		deploymentManifestParser:                deploymentManifestParser,
		tempRootConfigurator:                    tempRootConfigurator,
		targetProvider:                          targetProvider,
	}
}

type DeploymentPreparer struct {
	ui                                      biui.UI
	logger                                  boshlog.Logger
	logTag                                  string
	deploymentStateService                  biconfig.DeploymentStateService
	legacyDeploymentStateMigrator           biconfig.LegacyDeploymentStateMigrator
	releaseManager                          boshinst.ReleaseManager
	deploymentRecord                        bidepl.Record
	cloudFactory                            bicloud.Factory
	stemcellManagerFactory                  bistemcell.ManagerFactory
	agentClientFactory                      bihttpagent.AgentClientFactory
	vmManagerFactory                        bivm.ManagerFactory
	blobstoreFactory                        biblobstore.Factory
	deployer                                bidepl.Deployer
	deploymentManifestPath                  string
	deploymentVars                          boshtpl.Variables
	deploymentOp                            patch.Op
	cpiInstaller                            bicpirel.CpiInstaller
	releaseFetcher                          boshinst.ReleaseFetcher
	stemcellFetcher                         bistemcell.Fetcher
	releaseSetAndInstallationManifestParser ReleaseSetAndInstallationManifestParser
	deploymentManifestParser                DeploymentManifestParser
	tempRootConfigurator                    TempRootConfigurator
	targetProvider                          biinstall.TargetProvider
}

func (c *DeploymentPreparer) PrepareDeployment(stage biui.Stage, recreate bool, recreatePersistentDisks bool, skipDrain bool) (err error) {
	c.ui.BeginLinef("Deployment state: '%s'\n", c.deploymentStateService.Path())

	if !c.deploymentStateService.Exists() {
		migrated, err := c.legacyDeploymentStateMigrator.MigrateIfExists(biconfig.LegacyDeploymentStatePath(c.deploymentManifestPath))
		if err != nil {
			return bosherr.WrapError(err, "Migrating legacy deployment state file")
		}
		if migrated {
			c.ui.BeginLinef("Migrated legacy deployments file: '%s'\n", biconfig.LegacyDeploymentStatePath(c.deploymentManifestPath))
		}
	}

	deploymentState, err := c.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment state")
	}

	target, err := c.targetProvider.NewTarget()
	if err != nil {
		return bosherr.WrapError(err, "Determining installation target")
	}

	err = c.tempRootConfigurator.PrepareAndSetTempRoot(target.TmpPath(), c.logger)
	if err != nil {
		return bosherr.WrapError(err, "Setting temp root")
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
		manifestSHA          string
	)
	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		var releaseSetManifest birelsetmanifest.Manifest
		releaseSetManifest, installationManifest, err = c.releaseSetAndInstallationManifestParser.ReleaseSetAndInstallationManifest(c.deploymentManifestPath, c.deploymentVars, c.deploymentOp)
		if err != nil {
			return err
		}

		for _, releaseRef := range releaseSetManifest.Releases {
			err = c.releaseFetcher.DownloadAndExtract(releaseRef, stage)
			if err != nil {
				return err
			}
		}

		err := c.cpiInstaller.ValidateCpiRelease(installationManifest, stage)
		if err != nil {
			return err
		}

		deploymentManifest, manifestSHA, err = c.deploymentManifestParser.GetDeploymentManifest(c.deploymentManifestPath, c.deploymentVars, c.deploymentOp, releaseSetManifest, stage)
		if err != nil {
			return err
		}

		extractedStemcell, err = c.stemcellFetcher.GetStemcell(deploymentManifest, stage)
		return err
	})
	if err != nil {
		return err
	}
	defer func() {
		deleteErr := extractedStemcell.Cleanup()
		if deleteErr != nil {
			c.logger.Warn(c.logTag, "Failed to delete extracted stemcell: %s", deleteErr.Error())
		}
	}()

	isDeployed, err := c.deploymentRecord.IsDeployed(manifestSHA, c.releaseManager.List(), extractedStemcell)
	if err != nil {
		return bosherr.WrapError(err, "Checking if deployment has changed")
	}

	if isDeployed && !recreate && !recreatePersistentDisks {
		c.ui.BeginLinef("No deployment, stemcell or release changes. Skipping deploy.\n")
		return nil
	}

	err = c.cpiInstaller.WithInstalledCpiRelease(installationManifest, target, stage, func(installation biinstall.Installation) error {
		stemcellApiVersion := c.stemcellApiVersion(extractedStemcell)

		cloud, err := c.cloudFactory.NewCloud(installation, deploymentState.DirectorID, stemcellApiVersion)
		if err != nil {
			return bosherr.WrapError(err, "Creating CPI client from CPI installation")
		}

		deploy := func() error {
			return c.deploy(
				deploymentState,
				extractedStemcell,
				installationManifest,
				deploymentManifest,
				manifestSHA,
				skipDrain,
				stage,
				cloud,
			)
		}

		cpiInfo, err := cloud.Info()
		if err != nil {
			return bosherr.WrapError(err, "Error getting CPI info")
		}

		if stemcellApiVersion >= bicloud.StemcellNoRegistryAsOfVersion &&
			cpiInfo.ApiVersion == bicloud.MaxCpiApiVersionSupported {
			return deploy()
		} else {
			return bosherr.Errorf(
				"The `bosh` cli requires CPI v2.0 or greater, you are using %d",
				cpiInfo.ApiVersion,
			)
		}
	})
	return err
}

func (c *DeploymentPreparer) deploy(
	deploymentState biconfig.DeploymentState,
	extractedStemcell bistemcell.ExtractedStemcell,
	installationManifest biinstallmanifest.Manifest,
	deploymentManifest bideplmanifest.Manifest,
	manifestSHA string,
	skipDrain bool,
	stage biui.Stage,
	cloud bicloud.Cloud,
) (err error) {
	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)

	cloudStemcell, err := stemcellManager.Upload(extractedStemcell, stage)
	if err != nil {
		return err
	}

	agentClient, err := c.agentClientFactory.NewAgentClient(deploymentState.DirectorID, installationManifest.Mbus, installationManifest.Cert.CA)
	if err != nil {
		return err
	}
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient)

	blobstore, err := c.blobstoreFactory.Create(installationManifest.Mbus, bihttpclient.CreateDefaultClientInsecureSkipVerify())
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
			vmManager,
			blobstore,
			skipDrain,
			c.extractDiskCIDsFromState(deploymentState),
			deployStage,
		)
		if err != nil {
			return bosherr.WrapError(err, "Deploying")
		}

		err = c.deploymentRecord.Update(manifestSHA, c.releaseManager.List())
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

func (c *DeploymentPreparer) stemcellApiVersion(stemcell bistemcell.ExtractedStemcell) int {
	stemcellApiVersion := stemcell.Manifest().ApiVersion
	if stemcellApiVersion == 0 {
		return 1
	}
	return stemcellApiVersion
}

// These disk CIDs get passed all the way to the create_vm cpi call
func (c *DeploymentPreparer) extractDiskCIDsFromState(deploymentState biconfig.DeploymentState) []string {
	diskCIDs := make([]string, 0)
	for _, disk := range deploymentState.Disks {
		diskCIDs = append(diskCIDs, disk.CID)
	}

	return diskCIDs
}
