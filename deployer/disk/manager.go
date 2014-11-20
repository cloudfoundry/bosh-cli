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

	disk := NewDisk(diskRecord.CID, diskRecord.Size, diskRecord.CloudProperties)

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
			bosherr.WrapError(err,
				"Creating disk with size %s, cloudProperties %#v, instanceID %s",
				diskPool.DiskSize,
				diskCloudProperties,
				vmCID,
			)
	}

	_, err = m.diskRepo.Save(cid, diskPool.DiskSize, diskCloudProperties)
	if err != nil {
		return nil, bosherr.WrapError(err, "Saving deployment disk record")
	}

	disk := NewDisk(cid, diskPool.DiskSize, diskCloudProperties)

	return disk, nil
}
