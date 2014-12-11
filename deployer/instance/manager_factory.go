package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud, bmvm.Manager) Manager
}

type managerFactory struct {
	registryServerManager bmregistry.ServerManager
	sshTunnelFactory      bmsshtunnel.Factory
	diskDeployer          DiskDeployer
	logger                boshlog.Logger
}

func NewManagerFactory(
	registryServerManager bmregistry.ServerManager,
	sshTunnelFactory bmsshtunnel.Factory,
	diskDeployer DiskDeployer,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		registryServerManager: registryServerManager,
		sshTunnelFactory:      sshTunnelFactory,
		diskDeployer:          diskDeployer,
		logger:                logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, vmManager bmvm.Manager) Manager {
	return NewManager(
		cloud,
		vmManager,
		f.registryServerManager,
		f.sshTunnelFactory,
		f.diskDeployer,
		f.logger,
	)
}
