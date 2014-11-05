package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud) Manager
}

type managerFactory struct {
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	deploymentConfigService bmconfig.DeploymentConfigService
}

func NewManagerFactory(
	eventLogger bmeventlog.EventLogger,
	deploymentConfigService bmconfig.DeploymentConfigService,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		eventLogger:             eventLogger,
		deploymentConfigService: deploymentConfigService,
		logger:                  logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	return &manager{
		cloud:                   cloud,
		eventLogger:             f.eventLogger,
		deploymentConfigService: f.deploymentConfigService,
		logger:                  f.logger,
		logTag:                  "vmManager",
	}
}
