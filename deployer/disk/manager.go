package disk

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type Manager interface {
	FindCurrent() (Disk, bool, error)
	Create(bmdepl.DiskPool, string) (Disk, error)
	FindUnused() ([]Disk, error)
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

func (m *manager) Create(diskPool bmdepl.DiskPool, vmCID string) (Disk, error) {
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
