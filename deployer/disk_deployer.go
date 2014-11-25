package deployer

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type DiskDeployer interface {
	Deploy(diskPool bmdepl.DiskPool, cloud bmcloud.Cloud, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error
}

type diskDeployer struct {
	diskRepo           bmconfig.DiskRepo
	diskManagerFactory bmdisk.ManagerFactory
	diskManager        bmdisk.Manager
	eventLoggerStage   bmeventlog.Stage
	logger             boshlog.Logger
	logTag             string
}

func NewDiskDeployer(diskManagerFactory bmdisk.ManagerFactory, diskRepo bmconfig.DiskRepo, logger boshlog.Logger) DiskDeployer {
	return &diskDeployer{
		diskManagerFactory: diskManagerFactory,
		diskRepo:           diskRepo,
		logger:             logger,
		logTag:             "diskDeployer",
	}
}

func (d *diskDeployer) Deploy(diskPool bmdepl.DiskPool, cloud bmcloud.Cloud, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error {
	if diskPool.DiskSize == 0 {
		return nil
	}

	d.eventLoggerStage = eventLoggerStage

	d.diskManager = d.diskManagerFactory.NewManager(cloud)
	disk, diskFound, err := d.diskManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding existing disk")
	}

	if !diskFound {
		disk, err = d.createDisk(diskPool, vm)
		if err != nil {
			return err
		}
	}

	err = d.attachDisk(disk, vm)
	if err != nil {
		return err
	}

	if diskFound {
		diskCloudProperties, err := diskPool.CloudProperties()
		if err != nil {
			return bosherr.WrapError(err, "Getting disk pool cloud properties")
		}

		if disk.NeedsMigration(diskPool.DiskSize, diskCloudProperties) {
			disk, err = d.migrateDisk(disk, diskPool, vm)
			if err != nil {
				return err
			}
		}
	}

	err = d.updateCurrentDiskRecord(disk)
	if err != nil {
		return err
	}

	err = d.deleteUnusedDisks()
	if err != nil {
		return err
	}

	return nil
}

func (d *diskDeployer) migrateDisk(
	primaryDisk bmdisk.Disk,
	diskPool bmdepl.DiskPool,
	vm bmvm.VM,
) (bmdisk.Disk, error) {
	d.logger.Debug(d.logTag, "Migrating disk '%s'", primaryDisk.CID())

	createEventStep := d.eventLoggerStage.NewStep("Creating disk")
	createEventStep.Start()

	secondaryDisk, err := d.diskManager.Create(diskPool, vm.CID())
	if err != nil {
		err = bosherr.WrapError(err, "Creating secondary disk")
		createEventStep.Fail(err.Error())
		return nil, err
	}

	createEventStep.Finish()

	attachEventStep := d.eventLoggerStage.NewStep(fmt.Sprintf("Attaching disk '%s' to VM '%s'", secondaryDisk.CID(), vm.CID()))
	attachEventStep.Start()

	err = vm.AttachDisk(secondaryDisk)
	if err != nil {
		err = bosherr.WrapError(err, "Attaching secondary disk")
		attachEventStep.Fail(err.Error())
		return nil, err
	}
	attachEventStep.Finish()

	migrateEventStep := d.eventLoggerStage.NewStep(fmt.Sprintf("Migrating disk '%s' to '%s'", primaryDisk.CID(), secondaryDisk.CID()))
	migrateEventStep.Start()

	err = vm.MigrateDisk()
	if err != nil {
		err = bosherr.WrapError(err, "Migrating disk")
		migrateEventStep.Fail(err.Error())
		return nil, err
	}

	migrateEventStep.Finish()

	detachEventStep := d.eventLoggerStage.NewStep(fmt.Sprintf("Detaching disk '%s'", primaryDisk.CID()))
	detachEventStep.Start()

	err = vm.DetachDisk(primaryDisk)
	if err != nil {
		err = bosherr.WrapError(err, "Detaching disk")
		detachEventStep.Fail(err.Error())
		return nil, err
	}

	detachEventStep.Finish()

	return secondaryDisk, nil
}

func (d *diskDeployer) updateCurrentDiskRecord(disk bmdisk.Disk) error {
	savedDiskRecord, found, err := d.diskRepo.Find(disk.CID())
	if err != nil {
		return bosherr.WrapError(err, "Finding disk record")
	}

	if !found {
		return bosherr.New("Failed to find disk record for new disk")
	}

	err = d.diskRepo.UpdateCurrent(savedDiskRecord.ID)
	if err != nil {
		return bosherr.WrapError(err, "Updating current disk record")
	}

	return nil
}

func (d *diskDeployer) createDisk(diskPool bmdepl.DiskPool, vm bmvm.VM) (bmdisk.Disk, error) {
	d.logger.Debug(d.logTag, "Creating disk")

	createEventStep := d.eventLoggerStage.NewStep("Creating disk")
	createEventStep.Start()

	disk, err := d.diskManager.Create(diskPool, vm.CID())
	if err != nil {
		createEventStep.Fail(err.Error())
		return disk, bosherr.WrapError(err, "Creating new disk")
	}
	createEventStep.Finish()

	return disk, nil
}

func (d *diskDeployer) deleteUnusedDisks() error {
	disks, err := d.diskManager.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused disks")
	}

	for _, disk := range disks {
		deleteEventStep := d.eventLoggerStage.NewStep(fmt.Sprintf("Deleting unused disk '%s'", disk.CID()))
		deleteEventStep.Start()

		err = disk.Delete()
		if err != nil {
			err = bosherr.WrapError(err, "Deleting unused disk '%s'", disk.CID())
			deleteEventStep.Fail(err.Error())
			return err
		}
		deleteEventStep.Finish()
	}

	return nil
}

func (d *diskDeployer) attachDisk(disk bmdisk.Disk, vm bmvm.VM) error {
	d.logger.Debug(d.logTag, "Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID())
	attachEventStep := d.eventLoggerStage.NewStep(fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID()))
	attachEventStep.Start()

	err := vm.AttachDisk(disk)
	if err != nil {
		err = bosherr.WrapError(err, "Attaching disk")
		attachEventStep.Fail(err.Error())
		return err
	}

	attachEventStep.Finish()

	return nil
}
