package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud, bmvm.Manager, bmblobstore.Blobstore) Manager
}

type managerFactory struct {
	sshTunnelFactory bmsshtunnel.Factory
	instanceFactory  Factory
	logger           boshlog.Logger
}

func NewManagerFactory(
	sshTunnelFactory bmsshtunnel.Factory,
	instanceFactory Factory,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		sshTunnelFactory: sshTunnelFactory,
		instanceFactory:  instanceFactory,
		logger:           logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, vmManager bmvm.Manager, blobstore bmblobstore.Blobstore) Manager {
	return NewManager(
		cloud,
		vmManager,
		blobstore,
		f.sshTunnelFactory,
		f.instanceFactory,
		f.logger,
	)
}
