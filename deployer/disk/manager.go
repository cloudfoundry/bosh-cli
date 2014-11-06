package disk

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type Manager interface {
	Create(int, map[string]interface{}, string) (Disk, error)
}

type manager struct {
	cloud                   bmcloud.Cloud
	deploymentConfigService bmconfig.DeploymentConfigService
	logger                  boshlog.Logger
	logTag                  string
}

func (m *manager) Create(
	size int,
	cloudProperties map[string]interface{},
	instanceID string,
) (Disk, error) {
	m.logger.Debug(m.logTag, "Creating disk")
	cid, err := m.cloud.CreateDisk(size, cloudProperties, instanceID)
	if err != nil {
		return Disk{},
			bosherr.WrapError(err,
				"Creating disk with size %s, cloudProperties %#v, instanceID %s",
				size,
				cloudProperties,
				instanceID,
			)
	}

	deploymentConfig, err := m.deploymentConfigService.Load()
	if err != nil {
		return Disk{}, bosherr.WrapError(err, "Reading existing deployment config")
	}
	deploymentConfig.DiskCID = cid

	err = m.deploymentConfigService.Save(deploymentConfig)
	if err != nil {
		return Disk{}, bosherr.WrapError(err, "Saving deployment config")
	}

	disk := Disk{
		CID: cid,
	}

	return disk, nil
}
