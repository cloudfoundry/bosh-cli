package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
)

type ManagerFactory interface {
	NewManager(infrastructure Infrastructure) Manager
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

func (f *managerFactory) NewManager(infrastructure Infrastructure) Manager {
	return &manager{
		infrastructure:          infrastructure,
		eventLogger:             f.eventLogger,
		logger:                  f.logger,
		logTag:                  "vmManager",
		deploymentConfigService: f.deploymentConfigService,
	}
}
