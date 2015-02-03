package vm

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

// DiskDeployer is in the instance package to avoid a [disk -> vm -> disk] dependency cycle
type DiskDeployer interface {
	Deploy(diskPool bmdeplmanifest.DiskPool, cloud bmcloud.Cloud, vm VM, eventLoggerStage bmeventlog.Stage) ([]bmdisk.Disk, error)
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

func (d *diskDeployer) Deploy(diskPool bmdeplmanifest.DiskPool, cloud bmcloud.Cloud, vm VM, eventLoggerStage bmeventlog.Stage) ([]bmdisk.Disk, error) {
	if diskPool.DiskSize == 0 {
		return []bmdisk.Disk{}, nil
	}

	d.diskManager = d.diskManagerFactory.NewManager(cloud)
	disks, err := d.diskManager.FindCurrent()
	if err != nil {
		return disks, bosherr.WrapError(err, "Finding existing disk")
	}

	if len(disks) > 1 {
		return disks, bosherr.WrapError(err, "Multiple current disks not supported")

	} else if len(disks) == 1 {
		disks, err = d.deployExistingDisk(disks[0], diskPool, vm, eventLoggerStage)
		if err != nil {
			return disks, err
		}

	} else {
		disks, err = d.deployNewDisk(diskPool, vm, eventLoggerStage)
		if err != nil {
			return disks, err
		}
	}

	err = d.diskManager.DeleteUnused(eventLoggerStage)
	if err != nil {
		return disks, err
	}

	return disks, nil
}

func (d *diskDeployer) deployExistingDisk(disk bmdisk.Disk, diskPool bmdeplmanifest.DiskPool, vm VM, eventLoggerStage bmeventlog.Stage) ([]bmdisk.Disk, error) {
	disks := []bmdisk.Disk{}

	// the disk is already part of the deployment, and should already be attached
	disks = append(disks, disk)

	// attach is idempotent
	err := d.attachDisk(disk, vm, eventLoggerStage)
	if err != nil {
		return disks, err
	}

	if disk.NeedsMigration(diskPool.DiskSize, diskPool.CloudProperties) {
		disk, err = d.migrateDisk(disk, diskPool, vm, eventLoggerStage)
		if err != nil {
			return disks, err
		}

		// after migration, only the new disk is part of the deployment
		disks[0] = disk
	}

	return disks, nil
}

func (d *diskDeployer) deployNewDisk(diskPool bmdeplmanifest.DiskPool, vm VM, eventLoggerStage bmeventlog.Stage) ([]bmdisk.Disk, error) {
	disks := []bmdisk.Disk{}

	disk, err := d.createDisk(diskPool, vm, eventLoggerStage)
	if err != nil {
		return disks, err
	}

	err = d.attachDisk(disk, vm, eventLoggerStage)
	if err != nil {
		return disks, err
	}

	// once attached, the disk is part of the deployment
	disks = append(disks, disk)

	err = d.updateCurrentDiskRecord(disk)
	if err != nil {
		return disks, err
	}

	return disks, nil
}

func (d *diskDeployer) migrateDisk(
	originalDisk bmdisk.Disk,
	diskPool bmdeplmanifest.DiskPool,
	vm VM,
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

func (d *diskDeployer) createDisk(diskPool bmdeplmanifest.DiskPool, vm VM, eventLoggerStage bmeventlog.Stage) (disk bmdisk.Disk, err error) {
	err = eventLoggerStage.PerformStep("Creating disk", func() error {
		disk, err = d.diskManager.Create(diskPool, vm.CID())
		return err
	})

	return disk, err
}

func (d *diskDeployer) attachDisk(disk bmdisk.Disk, vm VM, eventLoggerStage bmeventlog.Stage) error {
	stepName := fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID())
	err := eventLoggerStage.PerformStep(stepName, func() error {
		return vm.AttachDisk(disk)
	})

	return err
}
