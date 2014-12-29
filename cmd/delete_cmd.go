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
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type deleteCmd struct {
	ui                      bmui.UI
	userConfig              bmconfig.UserConfig
	fs                      boshsys.FileSystem
	deploymentParser        bmmanifest.Parser
	deploymentConfigService bmconfig.DeploymentConfigService
	cpiDeploymentFactory    bmcpi.DeploymentFactory
	agentClientFactory      bmhttpagent.AgentClientFactory
	vmManagerFactory        bmvm.ManagerFactory
	instanceManagerFactory  bminstance.ManagerFactory
	diskManagerFactory      bmdisk.ManagerFactory
	stemcellManagerFactory  bmstemcell.ManagerFactory
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewDeleteCmd(ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	deploymentParser bmmanifest.Parser,
	deploymentConfigService bmconfig.DeploymentConfigService,
	cpiDeploymentFactory bmcpi.DeploymentFactory,
	agentClientFactory bmhttpagent.AgentClientFactory,
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	diskManagerFactory bmdisk.ManagerFactory,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger) *deleteCmd {
	return &deleteCmd{
		ui:                      ui,
		userConfig:              userConfig,
		fs:                      fs,
		deploymentParser:        deploymentParser,
		deploymentConfigService: deploymentConfigService,
		cpiDeploymentFactory:    cpiDeploymentFactory,
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

	var (
		cpiDeployment bmcpi.Deployment
	)
	err = validationStage.PerformStep("Validating deployment manifest", func() error {
		_, cpiDeploymentManifest, err := c.deploymentParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", deploymentManifestPath)
		}

		cpiDeployment = c.cpiDeploymentFactory.NewDeployment(cpiDeploymentManifest, deploymentConfig.DeploymentID, deploymentConfig.DirectorID)

		return nil
	})
	if err != nil {
		return err
	}

	var (
		cpiRelease bmrel.Release
	)
	err = validationStage.PerformStep("Validating cpi release", func() error {
		if !c.fs.FileExists(cpiReleaseTarballPath) {
			return bosherr.Errorf("Verifying that the CPI release '%s' exists", cpiReleaseTarballPath)
		}

		cpiRelease, err = cpiDeployment.ExtractRelease(cpiReleaseTarballPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting CPI release '%s'", cpiReleaseTarballPath)
		}

		return nil
	})
	if err != nil {
		return err
	}
	defer cpiRelease.Delete()

	validationStage.Finish()

	cloud, err := cpiDeployment.Install()
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI deployment")
	}

	err = cpiDeployment.StartJobs()
	if err != nil {
		return bosherr.WrapError(err, "Starting CPI jobs")
	}
	defer func() {
		err := cpiDeployment.StopJobs()
		c.logger.Warn(c.logTag, "CPI jobs failed to stop: %s", err)
	}()

	directorID := deploymentConfig.DirectorID
	mbusURL := cpiDeployment.Manifest().Mbus
	agentClient := c.agentClientFactory.NewAgentClient(directorID, mbusURL)
	vmManager := c.vmManagerFactory.NewManager(cloud, agentClient, mbusURL)
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
