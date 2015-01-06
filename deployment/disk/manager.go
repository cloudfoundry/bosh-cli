package disk

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Manager interface {
	FindCurrent() (Disk, bool, error)
	Create(bmdeplmanifest.DiskPool, string) (Disk, error)
	FindUnused() ([]Disk, error)
	DeleteUnused(bmeventlog.Stage) error
}

type manager struct {
	cloud    bmcloud.Cloud
	diskRepo bmconfig.DiskRepo
	logger   boshlog.Logger
	logTag   string
}

func (m *manager) FindCurrent() (Disk, bool, error) {
	diskRecord, found, err := m.diskRepo.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Reading disk record")
	}

	if !found {
		return nil, false, nil
	}

	disk := NewDisk(diskRecord, m.cloud, m.diskRepo)

	return disk, true, err
}

func (m *manager) Create(diskPool bmdeplmanifest.DiskPool, vmCID string) (Disk, error) {
	diskCloudProperties, err := diskPool.CloudProperties()
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading existing deployment config")
	}

	m.logger.Debug(m.logTag, "Creating disk")
	cid, err := m.cloud.CreateDisk(diskPool.DiskSize, diskCloudProperties, vmCID)
	if err != nil {
		return nil,
			bosherr.WrapErrorf(err,
				"Creating disk with size %d, cloudProperties %#v, instanceID %s",
				diskPool.DiskSize, diskCloudProperties, vmCID,
			)
	}

	diskRecord, err := m.diskRepo.Save(cid, diskPool.DiskSize, diskCloudProperties)
	if err != nil {
		return nil, bosherr.WrapError(err, "Saving deployment disk record")
	}

	disk := NewDisk(diskRecord, m.cloud, m.diskRepo)

	return disk, nil
}

func (m *manager) FindUnused() ([]Disk, error) {
	disks := []Disk{}

	diskRecords, err := m.diskRepo.All()
	if err != nil {
		return disks, bosherr.WrapError(err, "Getting all disk records")
	}

	currentDiskRecord, found, err := m.diskRepo.FindCurrent()
	if err != nil {
		return disks, bosherr.WrapError(err, "Finding current disk record")
	}

	for _, diskRecord := range diskRecords {
		if !found || diskRecord.ID != currentDiskRecord.ID {
			disks = append(disks, NewDisk(diskRecord, m.cloud, m.diskRepo))
		}
	}

	return disks, nil
}

func (m *manager) DeleteUnused(eventLoggerStage bmeventlog.Stage) error {
	disks, err := m.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused disks")
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Deleting unused disk '%s'", disk.CID())
		err = eventLoggerStage.PerformStep(stepName, func() error {
			err := disk.Delete()
			cloudErr, ok := err.(bmcloud.Error)
			if ok && cloudErr.Type() == bmcloud.DiskNotFoundError {
				return bmeventlog.NewSkippedStepError(cloudErr.Error())
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
