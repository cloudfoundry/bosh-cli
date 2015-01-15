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
	diskRepo bmconfig.DiskRepo
	logger   boshlog.Logger
}

func NewManagerFactory(
	diskRepo bmconfig.DiskRepo,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		diskRepo: diskRepo,
		logger:   logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	return NewManager(cloud, f.diskRepo, f.logger)
}
