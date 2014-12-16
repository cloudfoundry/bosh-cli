package instance

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

// DiskDeployer is in the instance package to avoid a [disk -> vm -> disk] dependency cycle
type DiskDeployer interface {
	Deploy(diskPool bmmanifest.DiskPool, cloud bmcloud.Cloud, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error
}

type diskDeployer struct {
	diskRepo           bmconfig.DiskRepo
	diskManagerFactory bmdisk.ManagerFactory
	diskManager        bmdisk.Manager
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

func (d *diskDeployer) Deploy(diskPool bmmanifest.DiskPool, cloud bmcloud.Cloud, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error {
	if diskPool.DiskSize == 0 {
		return nil
	}

	d.diskManager = d.diskManagerFactory.NewManager(cloud)
	disk, diskFound, err := d.diskManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding existing disk")
	}

	if !diskFound {
		disk, err = d.createDisk(diskPool, vm, eventLoggerStage)
		if err != nil {
			return err
		}
	}

	err = d.attachDisk(disk, vm, eventLoggerStage)
	if err != nil {
		return err
	}

	if !diskFound {
		err = d.updateCurrentDiskRecord(disk)
		if err != nil {
			return err
		}
	}

	if diskFound {
		diskCloudProperties, err := diskPool.CloudProperties()
		if err != nil {
			return bosherr.WrapError(err, "Getting disk pool cloud properties")
		}

		if disk.NeedsMigration(diskPool.DiskSize, diskCloudProperties) {
			disk, err = d.migrateDisk(disk, diskPool, vm, eventLoggerStage)
			if err != nil {
				return err
			}
		}
	}

	err = d.diskManager.DeleteUnused(eventLoggerStage)
	if err != nil {
		return err
	}

	return nil
}

func (d *diskDeployer) migrateDisk(
	originalDisk bmdisk.Disk,
	diskPool bmmanifest.DiskPool,
	vm bmvm.VM,
	eventLoggerStage bmeventlog.Stage,
) (newDisk bmdisk.Disk, err error) {
	d.logger.Debug(d.logTag, "Migrating disk '%s'", originalDisk.CID())

	err = eventLoggerStage.PerformStep("Creating disk", func() error {
		newDisk, err = d.diskManager.Create(diskPool, vm.CID())
		return err
	})
	if err != nil {
		return newDisk, err
	}

	stepName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", newDisk.CID(), vm.CID())
	err = eventLoggerStage.PerformStep(stepName, func() error {
		return vm.AttachDisk(newDisk)
	})
	if err != nil {
		return newDisk, err
	}

	stepName = fmt.Sprintf("Migrating disk content from '%s' to '%s'", originalDisk.CID(), newDisk.CID())
	err = eventLoggerStage.PerformStep(stepName, func() error {
		return vm.MigrateDisk()
	})
	if err != nil {
		return newDisk, err
	}

	err = d.updateCurrentDiskRecord(newDisk)
	if err != nil {
		return newDisk, err
	}

	stepName = fmt.Sprintf("Detaching disk '%s'", originalDisk.CID())
	err = eventLoggerStage.PerformStep(stepName, func() error {
		return vm.DetachDisk(originalDisk)
	})
	if err != nil {
		return newDisk, err
	}

	stepName = fmt.Sprintf("Deleting disk '%s'", originalDisk.CID())
	err = eventLoggerStage.PerformStep(stepName, func() error {
		return originalDisk.Delete()
	})
	if err != nil {
		return newDisk, err
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

func (d *diskDeployer) createDisk(diskPool bmmanifest.DiskPool, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) (disk bmdisk.Disk, err error) {
	err = eventLoggerStage.PerformStep("Creating disk", func() error {
		disk, err = d.diskManager.Create(diskPool, vm.CID())
		return err
	})

	return disk, err
}

func (d *diskDeployer) attachDisk(disk bmdisk.Disk, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error {
	stepName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID())
	err := eventLoggerStage.PerformStep(stepName, func() error {
		return vm.AttachDisk(disk)
	})

	return err
}
