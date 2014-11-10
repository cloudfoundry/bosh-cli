package disk

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud) Manager
}

type managerFactory struct {
	deploymentConfigService bmconfig.DeploymentConfigService
	logger                  boshlog.Logger
}

func NewManagerFactory(
	deploymentConfigService bmconfig.DeploymentConfigService,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		deploymentConfigService: deploymentConfigService,
		logger:                  logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	deploymentRecord := bmconfig.NewDeploymentRecord(f.deploymentConfigService, f.logger)

	return &manager{
		cloud:            cloud,
		deploymentRecord: deploymentRecord,
		logger:           f.logger,
		logTag:           "diskManager",
	}
}
