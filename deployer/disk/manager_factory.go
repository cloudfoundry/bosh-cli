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
	deploymentRecord bmconfig.DeploymentRecord
	logger           boshlog.Logger
}

func NewManagerFactory(
	deploymentRecord bmconfig.DeploymentRecord,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		deploymentRecord: deploymentRecord,
		logger:           logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	return &manager{
		cloud:            cloud,
		deploymentRecord: f.deploymentRecord,
		logger:           f.logger,
		logTag:           "diskManager",
	}
}
