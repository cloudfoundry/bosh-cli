package deployer

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type DiskDeployer interface {
	Deploy(diskPool bmdepl.DiskPool, cloud bmcloud.Cloud, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error
}

type diskDeployer struct {
	diskManagerFactory bmdisk.ManagerFactory
	logger             boshlog.Logger
	logTag             string
}

func NewDiskDeployer(diskManagerFactory bmdisk.ManagerFactory, logger boshlog.Logger) DiskDeployer {
	return &diskDeployer{
		diskManagerFactory: diskManagerFactory,
		logger:             logger,
		logTag:             "diskDeployer",
	}
}

func (d *diskDeployer) Deploy(diskPool bmdepl.DiskPool, cloud bmcloud.Cloud, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error {
	if diskPool.Size > 0 {
		d.logger.Debug(d.logTag, "Creating and Attaching disk to vm '%s'", vm.CID())
		diskManager := d.diskManagerFactory.NewManager(cloud)

		disk, found, err := diskManager.FindCurrent()
		if err != nil {
			return bosherr.WrapError(err, "Finding existing disk")
		}
		if !found {
			createEventStep := eventLoggerStage.NewStep("Creating disk")
			createEventStep.Start()

			disk, err = diskManager.Create(diskPool, vm.CID())
			if err != nil {
				createEventStep.Fail(err.Error())
				return bosherr.WrapError(err, "Creating new disk")
			}
			createEventStep.Finish()
		}

		attachEventStep := eventLoggerStage.NewStep(fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID()))
		attachEventStep.Start()

		err = vm.AttachDisk(disk)
		if err != nil {
			attachEventStep.Fail(err.Error())
			return err
		}
		attachEventStep.Finish()
	}

	return nil
}
