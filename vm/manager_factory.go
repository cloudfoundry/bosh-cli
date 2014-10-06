package vm

import (
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
)

type ManagerFactory interface {
	NewManager(infrastructure Infrastructure) Manager
}

type managerFactory struct {
	eventLogger             bmeventlog.EventLogger
	deploymentConfigService bmconfig.DeploymentConfigService
}

func NewManagerFactory(eventLogger bmeventlog.EventLogger, deploymentConfigService bmconfig.DeploymentConfigService) ManagerFactory {
	return &managerFactory{
		eventLogger:             eventLogger,
		deploymentConfigService: deploymentConfigService,
	}
}

func (f *managerFactory) NewManager(infrastructure Infrastructure) Manager {
	return &manager{
		infrastructure:          infrastructure,
		eventLogger:             f.eventLogger,
		deploymentConfigService: f.deploymentConfigService,
	}
}
