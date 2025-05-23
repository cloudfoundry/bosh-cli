package deployment

import (
	"errors"
	"fmt"
	"time"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type Deployment interface {
	Delete(bool, biui.Stage) error
	Stop(bool, biui.Stage) error
	Start(biui.Stage, bideplmanifest.Update) error
}

type deployment struct {
	instances   []biinstance.Instance
	disks       []bidisk.Disk
	stemcells   []bistemcell.CloudStemcell
	pingTimeout time.Duration
	pingDelay   time.Duration
}

func NewDeployment(
	instances []biinstance.Instance,
	disks []bidisk.Disk,
	stemcells []bistemcell.CloudStemcell,
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

func (d *deployment) Delete(skipDrain bool, deleteStage biui.Stage) error {
	// le sigh... consuming from an array sucks without generics
	for len(d.instances) > 0 {
		lastIdx := len(d.instances) - 1
		instance := d.instances[lastIdx]

		if err := instance.Delete(d.pingTimeout, d.pingDelay, skipDrain, deleteStage); err != nil {
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

func (d *deployment) Stop(skipDrain bool, stopEnvStage biui.Stage) error {
	for len(d.instances) > 0 {
		lastIdx := len(d.instances) - 1
		instance := d.instances[lastIdx]

		if err := instance.Stop(d.pingTimeout, d.pingDelay, skipDrain, stopEnvStage); err != nil {
			return err
		}

		d.instances = d.instances[:lastIdx]
	}

	return nil
}

func (d *deployment) Start(startEnvStage biui.Stage, updateSection bideplmanifest.Update) error {
	for len(d.instances) > 0 {
		lastIdx := len(d.instances) - 1
		instance := d.instances[lastIdx]

		if err := instance.Start(updateSection, d.pingTimeout, d.pingDelay, startEnvStage); err != nil {
			return err
		}

		d.instances = d.instances[:lastIdx]
	}

	return nil
}

func (d *deployment) deleteDisk(deleteStage biui.Stage, disk bidisk.Disk) error {
	stepName := fmt.Sprintf("Deleting disk '%s'", disk.CID())
	return deleteStage.Perform(stepName, func() error {
		err := disk.Delete()
		var cloudErr bicloud.Error
		ok := errors.As(err, &cloudErr)
		if ok && cloudErr.Type() == bicloud.DiskNotFoundError {
			return biui.NewSkipStageError(cloudErr, "Disk not found")
		}
		return err
	})
}

func (d *deployment) deleteStemcell(deleteStage biui.Stage, stemcell bistemcell.CloudStemcell) error {
	stepName := fmt.Sprintf("Deleting stemcell '%s'", stemcell.CID())
	return deleteStage.Perform(stepName, func() error {
		err := stemcell.Delete()
		var cloudErr bicloud.Error
		ok := errors.As(err, &cloudErr)
		if ok && cloudErr.Type() == bicloud.StemcellNotFoundError {
			return biui.NewSkipStageError(cloudErr, "Stemcell not found")
		}
		return err
	})
}
