package cmd

import (
	"errors"
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpirel "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type deleteCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	installationParser      bminstallmanifest.Parser
	deploymentConfigService bmconfig.DeploymentConfigService
	installerFactory        bminstall.InstallerFactory
	releaseExtractor        bmrel.Extractor
	releaseManager          bmrel.Manager
	cloudFactory            bmcloud.Factory
	agentClientFactory      bmhttpagent.AgentClientFactory
	vmManagerFactory        bmvm.ManagerFactory
	instanceManagerFactory  bminstance.ManagerFactory
	diskManagerFactory      bmdisk.ManagerFactory
	stemcellManagerFactory  bmstemcell.ManagerFactory
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewDeleteCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	installationParser bminstallmanifest.Parser,
	deploymentConfigService bmconfig.DeploymentConfigService,
	installerFactory bminstall.InstallerFactory,
	releaseExtractor bmrel.Extractor,
	releaseManager bmrel.Manager,
	cloudFactory bmcloud.Factory,
	agentClientFactory bmhttpagent.AgentClientFactory,
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	diskManagerFactory bmdisk.ManagerFactory,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deleteCmd {
	return &deleteCmd{
		ui:                      ui,
		userConfig:              userConfig,
		fs:                      fs,
		installationParser:      installationParser,
		deploymentConfigService: deploymentConfigService,
		installerFactory:        installerFactory,
		releaseExtractor:        releaseExtractor,
		releaseManager:          releaseManager,
		cloudFactory:            cloudFactory,
		agentClientFactory:      agentClientFactory,
		vmManagerFactory:        vmManagerFactory,
		instanceManagerFactory:  instanceManagerFactory,
		diskManagerFactory:      diskManagerFactory,
		stemcellManagerFactory:  stemcellManagerFactory,
		eventLogger:             eventLogger,
		logger:                  logger,
		logTag:                  "deleteCmd",
	}
}

func (c *deleteCmd) Name() string {
	return "delete"
}

func (c *deleteCmd) Run(args []string) error {
	cpiReleaseTarballPath, err := c.parseCmdInputs(args)
	if err != nil {
		return err
	}

	deploymentManifestPath, err := getDeploymentManifest(c.userConfig, c.ui, c.fs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running delete cmd")
	}

	validationStage := c.eventLogger.NewStage("validating")
	validationStage.Start()

	deploymentConfig, err := c.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}

	var installationManifest bminstallmanifest.Manifest
	err = validationStage.PerformStep("Validating deployment manifest", func() error {
		var err error
		installationManifest, err = c.installationParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing installation manifest '%s'", deploymentManifestPath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	var cpiRelease bmrel.Release
	err = validationStage.PerformStep("Validating cpi release", func() error {
		if !c.fs.FileExists(cpiReleaseTarballPath) {
			return bosherr.Errorf("Verifying that the CPI release '%s' exists", cpiReleaseTarballPath)
		}

		var err error
		cpiRelease, err = c.releaseExtractor.Extract(cpiReleaseTarballPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting CPI release '%s'", cpiReleaseTarballPath)
		}
		c.releaseManager.Add(cpiRelease)

		err = bmcpirel.NewCpiValidator().Validate(cpiRelease)
		if err != nil {
			return bosherr.WrapError(err, "Invalid CPI release")
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

	validationStage.Finish()

	releaseResolver := bmrel.NewResolver(c.logger, c.releaseManager, []bmdeplmanifest.ReleaseRef{})
	installer, err := c.installerFactory.NewInstaller(releaseResolver)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI Installer")
	}

	installation, err := installer.Install(installationManifest)
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI")
	}

	err = installation.StartRegistry()
	if err != nil {
		return bosherr.WrapError(err, "Starting Registry")
	}
	defer func() {
		err := installation.StopRegistry()
		if err != nil {
			c.logger.Warn(c.logTag, "Registry failed to stop: %s", err)
		}
	}()

	cloud, err := c.cloudFactory.NewCloud(installation, deploymentConfig.DirectorID)
	if err != nil {
		return bosherr.WrapError(err, "Creating CPI client from CPI installation")
	}

	agentClient := c.agentClientFactory.NewAgentClient(deploymentConfig.DirectorID, installationManifest.Mbus)
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient, installationManifest.Mbus)
	instanceManager := c.instanceManagerFactory.NewManager(cloud, vmManager)
	diskManager := c.diskManagerFactory.NewManager(cloud)
	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)

	//TODO: deployment.Delete()
	return c.deleteDeployment(
		instanceManager,
		diskManager,
		stemcellManager,
	)
}

func (c *deleteCmd) parseCmdInputs(args []string) (string, error) {
	if len(args) != 1 {
		c.ui.Error("Invalid usage - delete command requires exactly 1 argument")
		c.ui.Sayln("Expected usage: bosh-micro delete <cpi-release-tarball>")
		c.logger.Error(c.logTag, "Invalid arguments: %#v", args)
		return "", errors.New("Invalid usage - delete command requires exactly 1 argument")
	}
	return args[0], nil
}

func (c *deleteCmd) deleteDisk(deleteStage bmeventlog.Stage, disk bmdisk.Disk) error {
	stepName := fmt.Sprintf("Deleting disk '%s'", disk.CID())
	return deleteStage.PerformStep(stepName, func() error {
		err := disk.Delete()
		cloudErr, ok := err.(bmcloud.Error)
		if ok && cloudErr.Type() == bmcloud.DiskNotFoundError {
			return bmeventlog.NewSkippedStepError(cloudErr.Error())
		}
		return err
	})
}

func (c *deleteCmd) deleteStemcell(deleteStage bmeventlog.Stage, stemcell bmstemcell.CloudStemcell) error {
	stepName := fmt.Sprintf("Deleting stemcell '%s'", stemcell.CID())
	return deleteStage.PerformStep(stepName, func() error {
		err := stemcell.Delete()
		cloudErr, ok := err.(bmcloud.Error)
		if ok && cloudErr.Type() == bmcloud.StemcellNotFoundError {
			return bmeventlog.NewSkippedStepError(cloudErr.Error())
		}
		return err
	})
}

func (c *deleteCmd) deleteDeployment(
	instanceManager bminstance.Manager,
	diskManager bmdisk.Manager,
	stemcellManager bmstemcell.Manager,
) error {
	deleteStage := c.eventLogger.NewStage("deleting deployment")
	deleteStage.Start()

	instances, err := instanceManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment instances")
	}

	disk, diskFound, err := diskManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment disk")
	}

	stemcell, stemcellFound, err := stemcellManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment stemcell")
	}

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond
	for _, instance := range instances {
		if err = instance.Delete(pingTimeout, pingDelay, deleteStage); err != nil {
			return err
		}
	}

	if diskFound {
		if err = c.deleteDisk(deleteStage, disk); err != nil {
			return err
		}
	}

	if stemcellFound {
		if err = c.deleteStemcell(deleteStage, stemcell); err != nil {
			return err
		}
	}

	if err = diskManager.DeleteUnused(deleteStage); err != nil {
		return err
	}

	if err = stemcellManager.DeleteUnused(deleteStage); err != nil {
		return err
	}

	deleteStage.Finish()

	return nil
}
