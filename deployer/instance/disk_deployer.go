package instance

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

// DiskDeployer is in the instance package to avoid a [disk -> vm -> disk] dependency cycle
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
	originalDisk bmdisk.Disk,
	diskPool bmdepl.DiskPool,
	vm bmvm.VM,
) (newDisk bmdisk.Disk, err error) {
	d.logger.Debug(d.logTag, "Migrating disk '%s'", originalDisk.CID())

	err = d.eventLoggerStage.PerformStep("Creating disk", func() error {
		newDisk, err = d.diskManager.Create(diskPool, vm.CID())
		if err != nil {
			return bosherr.WrapError(err, "Creating secondary disk")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	stepName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", newDisk.CID(), vm.CID())
	err = d.eventLoggerStage.PerformStep(stepName, func() error {
		if err = vm.AttachDisk(newDisk); err != nil {
			return bosherr.WrapError(err, "Attaching secondary disk")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	stepName = fmt.Sprintf("Migrating disk '%s' to '%s'", originalDisk.CID(), newDisk.CID())
	err = d.eventLoggerStage.PerformStep(stepName, func() error {
		if err = vm.MigrateDisk(); err != nil {
			return bosherr.WrapError(err, "Migrating disk")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	stepName = fmt.Sprintf("Detaching disk '%s'", originalDisk.CID())
	err = d.eventLoggerStage.PerformStep(stepName, func() error {
		if err = vm.DetachDisk(originalDisk); err != nil {
			return bosherr.WrapError(err, "Detaching disk")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return newDisk, nil
}

func (d *diskDeployer) updateCurrentDiskRecord(disk bmdisk.Disk) error {
	savedDiskRecord, found, err := d.diskRepo.Find(disk.CID())
	if err != nil {
		return bosherr.WrapError(err, "Finding disk record")
	}

	if !found {
		return bosherr.Error("Failed to find disk record for new disk")
	}

	err = d.diskRepo.UpdateCurrent(savedDiskRecord.ID)
	if err != nil {
		return bosherr.WrapError(err, "Updating current disk record")
	}

	return nil
}

func (d *diskDeployer) createDisk(diskPool bmdepl.DiskPool, vm bmvm.VM) (disk bmdisk.Disk, err error) {
	err = d.eventLoggerStage.PerformStep("Creating disk", func() error {
		disk, err = d.diskManager.Create(diskPool, vm.CID())
		if err != nil {
			return bosherr.WrapError(err, "Creating new disk")
		}
		return nil
	})

	return disk, err
}

func (d *diskDeployer) deleteUnusedDisks() error {
	disks, err := d.diskManager.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused disks")
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Deleting unused disk '%s'", disk.CID())
		err = d.eventLoggerStage.PerformStep(stepName, func() error {
			err = disk.Delete()
			if err != nil {
				return bosherr.WrapErrorf(err, "Deleting unused disk '%s'", disk.CID())
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *diskDeployer) attachDisk(disk bmdisk.Disk, vm bmvm.VM) error {
	stepName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID())
	err := d.eventLoggerStage.PerformStep(stepName, func() error {
		err := vm.AttachDisk(disk)
		if err != nil {
			return bosherr.WrapError(err, "Attaching disk")
		}

		return nil
	})

	return err
}
