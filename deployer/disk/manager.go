package disk

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type Manager interface {
	Create(bmdepl.DiskPool, string) (Disk, error)
}

type manager struct {
	cloud                   bmcloud.Cloud
	deploymentConfigService bmconfig.DeploymentConfigService
	logger                  boshlog.Logger
	logTag                  string
}

func (m *manager) Create(diskPool bmdepl.DiskPool, instanceID string) (Disk, error) {
	deploymentConfig, err := m.deploymentConfigService.Load()
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading existing deployment config")
	}

	if cid := deploymentConfig.DiskCID; cid != "" {
		m.logger.Debug(m.logTag, "Using existing disk '%s'", cid)
		disk := NewDisk(cid)
		return disk, nil
	}

	diskCloudProperties, err := diskPool.CloudProperties()
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading existing deployment config")
	}

	m.logger.Debug(m.logTag, "Creating disk")
	cid, err := m.cloud.CreateDisk(diskPool.Size, diskCloudProperties, instanceID)
	if err != nil {
		return nil,
			bosherr.WrapError(err,
				"Creating disk with size %s, cloudProperties %#v, instanceID %s",
				diskPool.Size,
				diskCloudProperties,
				instanceID,
			)
	}

	deploymentConfig.DiskCID = cid

	err = m.deploymentConfigService.Save(deploymentConfig)
	if err != nil {
		return nil, bosherr.WrapError(err, "Saving deployment config")
	}

	disk := NewDisk(cid)

	return disk, nil
}
