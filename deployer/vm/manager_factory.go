package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud) Manager
}

type managerFactory struct {
	logger                  boshlog.Logger
	deploymentConfigService bmconfig.DeploymentConfigService
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
	return &manager{
		cloud: cloud,
		deploymentConfigService: f.deploymentConfigService,
		logger:                  f.logger,
		logTag:                  "vmManager",
	}
}
