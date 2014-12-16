package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud, bmvm.Manager) Manager
}

type managerFactory struct {
	sshTunnelFactory bmsshtunnel.Factory
	diskDeployer     DiskDeployer
	logger           boshlog.Logger
}

func NewManagerFactory(
	sshTunnelFactory bmsshtunnel.Factory,
	diskDeployer DiskDeployer,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		sshTunnelFactory: sshTunnelFactory,
		diskDeployer:     diskDeployer,
		logger:           logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, vmManager bmvm.Manager) Manager {
	return NewManager(
		cloud,
		vmManager,
		f.sshTunnelFactory,
		f.diskDeployer,
		f.logger,
	)
}
