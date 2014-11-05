package disk

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud) Manager
}

type managerFactory struct {
	logger boshlog.Logger
}

func NewManagerFactory(logger boshlog.Logger) ManagerFactory {
	return &managerFactory{
		logger: logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	return &manager{
		cloud:  cloud,
		logger: f.logger,
		logTag: "diskManager",
	}
}
