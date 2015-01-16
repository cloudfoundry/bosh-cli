package stemcell

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud) Manager
}

type managerFactory struct {
	repo bmconfig.StemcellRepo
}

func NewManagerFactory(repo bmconfig.StemcellRepo) ManagerFactory {
	return &managerFactory{
		repo: repo,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	return NewManager(f.repo, cloud)
}
