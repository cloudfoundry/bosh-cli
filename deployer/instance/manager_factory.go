package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud, bmvm.Manager) Manager
}

type managerFactory struct {
	registryServer   bmregistry.Server
	sshTunnelFactory bmsshtunnel.Factory
	diskDeployer     DiskDeployer
	logger           boshlog.Logger
}

func NewManagerFactory(
	registryServer bmregistry.Server,
	sshTunnelFactory bmsshtunnel.Factory,
	diskDeployer DiskDeployer,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		registryServer:   registryServer,
		sshTunnelFactory: sshTunnelFactory,
		diskDeployer:     diskDeployer,
		logger:           logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, vmManager bmvm.Manager) Manager {
	return NewManager(
		cloud,
		vmManager,
		f.registryServer,
		f.sshTunnelFactory,
		f.diskDeployer,
		f.logger,
	)
}
