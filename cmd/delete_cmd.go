package cmd

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type deleteCmd struct {
	ui                 bmui.UI
	userConfig         bmconfig.UserConfig
	fs                 boshsys.FileSystem
	deploymentParser   bmdepl.Parser
	cpiInstaller       bmcpi.Installer
	vmRepo             bmconfig.VMRepo
	diskRepo           bmconfig.DiskRepo
	stemcellRepo       bmconfig.StemcellRepo
	agentClientFactory bmagentclient.Factory
	eventLogger        bmeventlog.EventLogger
	logger             boshlog.Logger
	logTag             string
}

func NewDeleteCmd(ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	deploymentParser bmdepl.Parser,
	cpiInstaller bmcpi.Installer,
	vmRepo bmconfig.VMRepo,
	diskRepo bmconfig.DiskRepo,
	stemcellRepo bmconfig.StemcellRepo,
	agentClientFactory bmagentclient.Factory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger) *deleteCmd {
	return &deleteCmd{
		ui:                 ui,
		userConfig:         userConfig,
		fs:                 fs,
		deploymentParser:   deploymentParser,
		cpiInstaller:       cpiInstaller,
		vmRepo:             vmRepo,
		diskRepo:           diskRepo,
		stemcellRepo:       stemcellRepo,
		agentClientFactory: agentClientFactory,
		eventLogger:        eventLogger,
		logger:             logger,
		logTag:             "deleteCmd",
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

	cpiDeployment, cpiRelease, err := c.validateInputFiles(cpiReleaseTarballPath)
	if err != nil {
		return err
	}
	defer cpiRelease.Delete()

	cloud, err := c.cpiInstaller.Install(cpiDeployment, cpiRelease)
	if err != nil {
		return bosherr.WrapError(err, "Installing CPI deployment")
	}

	deleteStage := c.eventLogger.NewStage("deleting deployment")
	deleteStage.Start()

	vmCID, vmFound, err := c.vmRepo.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment VM")
	}

	diskManager := bmdisk.NewManagerFactory(c.diskRepo, c.logger).NewManager(cloud)
	disk, diskFound, err := diskManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment disk")
	}

	stemcellManager := bmstemcell.NewManagerFactory(c.stemcellRepo, c.eventLogger).NewManager(cloud)
	stemcell, stemcellFound, err := stemcellManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding current deployment stemcell")
	}

	if !vmFound && !diskFound && !stemcellFound {
		c.ui.Error("No existing microbosh instance to delete")
		return bosherr.New("No existing microbosh instance to delete")
	}

	stopAgentStep := deleteStage.NewStep("Stopping agent")
	stopAgentStep.Start()
	agentClient := c.agentClientFactory.Create(cpiDeployment.Mbus)
	err = agentClient.Stop()
	if err != nil {
		err = bosherr.WrapError(err, "Stopping the agent with mbus `%s'", cpiDeployment.Mbus)
		stopAgentStep.Fail(err.Error())
		return err
	}
	stopAgentStep.Finish()

	if vmFound {
		deleteVMStep := deleteStage.NewStep("Deleting VM")
		deleteVMStep.Start()
		err = cloud.DeleteVM(vmCID)
		if err != nil {
			err = bosherr.WrapError(err, "Deleting deployment VM `%s'", vmCID)
			deleteVMStep.Fail(err.Error())
			return err
		}
		err = c.vmRepo.ClearCurrent()
		if err != nil {
			err = bosherr.WrapError(err, "Deleting deployment VM record `%s'", vmCID)
			deleteVMStep.Fail(err.Error())
			return err
		}
		deleteVMStep.Finish()
	}

	if diskFound {
		deleteDiskStep := deleteStage.NewStep("Deleting disk")
		deleteDiskStep.Start()
		err = disk.Delete()
		if err != nil {
			err = bosherr.WrapError(err, "Deleting deployment disk `%s'", disk.CID())
			deleteDiskStep.Fail(err.Error())
			return err
		}
		if err = c.diskRepo.ClearCurrent(); err != nil {
			err = bosherr.WrapError(err, "Deleting deployment disk record `%s'", disk.CID())
			deleteDiskStep.Fail(err.Error())
			return err
		}
		deleteDiskStep.Finish()
	}

	if stemcellFound {
		deleteStemcellStep := deleteStage.NewStep("Deleting stemcell")
		deleteStemcellStep.Start()
		err = stemcell.Delete()
		if err != nil {
			err = bosherr.WrapError(err, "Deleting deployment stemcell `%s'", stemcell.CID())
			deleteStemcellStep.Fail(err.Error())
			return err
		}
		if err = c.stemcellRepo.ClearCurrent(); err != nil {
			err = bosherr.WrapError(err, "Deleting deployment stemcell record `%s'", stemcell.CID())
			deleteStemcellStep.Fail(err.Error())
			return err
		}
		deleteStemcellStep.Finish()
	}

	unusedDisks, err := diskManager.FindUnused()
	if len(unusedDisks) > 0 {
		deleteDiskStep := deleteStage.NewStep("Deleting orphaned disks")
		deleteDiskStep.Start()
		for _, disk := range unusedDisks {
			err = disk.Delete()
			if err != nil {
				err = bosherr.WrapError(err, "Deleting orphaned disk `%s'", disk.CID())
				deleteDiskStep.Fail(err.Error())
				return err
			}
		}
		deleteDiskStep.Finish()
	}

	unusedStemcells, err := stemcellManager.FindUnused()
	if len(unusedStemcells) > 0 {
		deleteStemcellStep := deleteStage.NewStep("Deleting orphaned stemcells")
		deleteStemcellStep.Start()
		for _, stemcell := range unusedStemcells {
			err = stemcell.Delete()
			if err != nil {
				err = bosherr.WrapError(err, "Deleting orphaned stemcell `%s'", stemcell.CID())
				deleteStemcellStep.Fail(err.Error())
				return err
			}
		}
		deleteStemcellStep.Finish()
	}

	deleteStage.Finish()

	return nil
}

func (c *deleteCmd) validateInputFiles(releaseTarballPath string) (
	cpiDeployment bmdepl.CPIDeployment,
	cpiRelease bmrel.Release,
	err error,
) {
	validationStage := c.eventLogger.NewStage("validating")
	validationStage.Start()

	manifestValidationStep := validationStage.NewStep("Validating deployment manifest")
	manifestValidationStep.Start()

	if c.userConfig.DeploymentFile == "" {
		err = bosherr.New("No deployment set")
		manifestValidationStep.Fail(err.Error())
		return cpiDeployment, nil, err
	}

	deploymentFilePath := c.userConfig.DeploymentFile

	c.logger.Info(c.logTag, "Checking for deployment `%s'", deploymentFilePath)
	if !c.fs.FileExists(deploymentFilePath) {
		err = bosherr.New("Verifying that the deployment `%s' exists", deploymentFilePath)
		manifestValidationStep.Fail(err.Error())
		return cpiDeployment, nil, err
	}

	_, cpiDeployment, err = c.deploymentParser.Parse(deploymentFilePath)
	if err != nil {
		err = bosherr.WrapError(err, "Parsing deployment manifest `%s'", deploymentFilePath)
		manifestValidationStep.Fail(err.Error())
		return cpiDeployment, nil, err
	}

	manifestValidationStep.Finish()

	cpiValidationStep := validationStage.NewStep("Validating cpi release")
	cpiValidationStep.Start()

	if !c.fs.FileExists(releaseTarballPath) {
		err = bosherr.New("Verifying that the CPI release `%s' exists", releaseTarballPath)
		cpiValidationStep.Fail(err.Error())
		return cpiDeployment, cpiRelease, err
	}

	cpiRelease, err = c.cpiInstaller.Extract(releaseTarballPath)
	if err != nil {
		err = bosherr.WrapError(err, "Extracting CPI release `%s'", releaseTarballPath)
		cpiValidationStep.Fail(err.Error())
		return cpiDeployment, cpiRelease, err
	}

	cpiValidationStep.Finish()

	validationStage.Finish()

	return cpiDeployment, cpiRelease, nil
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
