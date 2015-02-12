package deployment

import (
	"fmt"
	"time"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Deployment interface {
	Delete(bmui.Stage) error
}

type deployment struct {
	instances   []bminstance.Instance
	disks       []bmdisk.Disk
	stemcells   []bmstemcell.CloudStemcell
	pingTimeout time.Duration
	pingDelay   time.Duration
}

func NewDeployment(
	instances []bminstance.Instance,
	disks []bmdisk.Disk,
	stemcells []bmstemcell.CloudStemcell,
	pingTimeout time.Duration,
	pingDelay time.Duration,
) Deployment {
	return &deployment{
		instances:   instances,
		disks:       disks,
		stemcells:   stemcells,
		pingTimeout: pingTimeout,
		pingDelay:   pingDelay,
	}
}

func (d *deployment) Delete(deleteStage bmui.Stage) error {
	// le sigh... consuming from an array sucks without generics
	for len(d.instances) > 0 {
		lastIdx := len(d.instances) - 1
		instance := d.instances[lastIdx]

		if err := instance.Delete(d.pingTimeout, d.pingDelay, deleteStage); err != nil {
			return err
		}

		d.instances = d.instances[:lastIdx]
	}

	for len(d.disks) > 0 {
		lastIdx := len(d.disks) - 1
		disk := d.disks[lastIdx]

		if err := d.deleteDisk(deleteStage, disk); err != nil {
			return err
		}

		d.disks = d.disks[:lastIdx]
	}

	for len(d.stemcells) > 0 {
		lastIdx := len(d.stemcells) - 1
		stemcell := d.stemcells[lastIdx]

		if err := d.deleteStemcell(deleteStage, stemcell); err != nil {
			return err
		}

		d.stemcells = d.stemcells[:lastIdx]
	}

	return nil
}

func (d *deployment) deleteDisk(deleteStage bmui.Stage, disk bmdisk.Disk) error {
	stepName := fmt.Sprintf("Deleting disk '%s'", disk.CID())
	return deleteStage.Perform(stepName, func() error {
		err := disk.Delete()
		cloudErr, ok := err.(bmcloud.Error)
		if ok && cloudErr.Type() == bmcloud.DiskNotFoundError {
			return bmui.NewSkipStageError(cloudErr, "Disk not found")
		}
		return err
	})
}

func (d *deployment) deleteStemcell(deleteStage bmui.Stage, stemcell bmstemcell.CloudStemcell) error {
	stepName := fmt.Sprintf("Deleting stemcell '%s'", stemcell.CID())
	return deleteStage.Perform(stepName, func() error {
		err := stemcell.Delete()
		cloudErr, ok := err.(bmcloud.Error)
		if ok && cloudErr.Type() == bmcloud.StemcellNotFoundError {
			return bmui.NewSkipStageError(cloudErr, "Stemcell not found")
		}
		return err
	})
}
