package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type DiskRecord struct {
	CID             string
	Size            int
	CloudProperties map[string]interface{}
}

type DeploymentRecord interface {
	Disk() (DiskRecord, bool, error)
	UpdateDisk(DiskRecord) error
}

type deploymentRecord struct {
	configService DeploymentConfigService
	logger        boshlog.Logger
	logTag        string
}

func NewDeploymentRecord(
	configService DeploymentConfigService,
	logger boshlog.Logger,
) DeploymentRecord {
	return &deploymentRecord{
		configService: configService,
		logger:        logger,
		logTag:        "deploymentRecord",
	}
}

func (dr *deploymentRecord) Disk() (DiskRecord, bool, error) {
	deploymentConfig, err := dr.configService.Load()
	if err != nil {
		return DiskRecord{}, false, bosherr.WrapError(err, "Reading deployment disk record")
	}

	if deploymentConfig.DiskCID == "" {
		return DiskRecord{}, false, nil
	}

	return DiskRecord{
		CID: deploymentConfig.DiskCID,
	}, true, nil
}

func (dr *deploymentRecord) UpdateDisk(disk DiskRecord) error {
	deploymentConfig, err := dr.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Reading deployment disk record for update")
	}

	deploymentConfig.DiskCID = disk.CID

	err = dr.configService.Save(deploymentConfig)
	if err != nil {
		return bosherr.WrapError(err, "Updating deployment disk record")
	}

	return nil
}
