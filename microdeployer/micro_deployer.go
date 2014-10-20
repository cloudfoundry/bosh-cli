package microdeployer

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type Deployer interface {
	Deploy(bmcloud.Cloud, bmdepl.Deployment, bmdepl.Registry, bmdepl.SSHTunnel, bmstemcell.CID) error
}

type microDeployer struct {
	vmManagerFactory bmvm.ManagerFactory
	sshTunnelFactory bmsshtunnel.Factory
	registryServer   bmregistry.Server
	logger           boshlog.Logger
	logTag           string
}

func NewMicroDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	registryServer bmregistry.Server,
	logger boshlog.Logger,
) *microDeployer {
	return &microDeployer{
		vmManagerFactory: vmManagerFactory,
		sshTunnelFactory: sshTunnelFactory,
		registryServer:   registryServer,
		logger:           logger,
		logTag:           "microDeployer",
	}
}

func (m *microDeployer) Deploy(
	cpi bmcloud.Cloud,
	deployment bmdepl.Deployment,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	stemcellCID bmstemcell.CID,
) error {
	registryReadyCh := make(chan struct{})
	go m.startRegistry(registry, registryReadyCh)
	defer m.registryServer.Stop()

	<-registryReadyCh

	vmManager := m.vmManagerFactory.NewManager(cpi)
	_, err := vmManager.CreateVM(stemcellCID, deployment)
	if err != nil {
		return bosherr.WrapError(err, "Creating VM")
	}

	sshTunnelOptions := bmsshtunnel.Options{
		Host:              sshTunnelConfig.Host,
		Port:              sshTunnelConfig.Port,
		User:              sshTunnelConfig.User,
		Password:          sshTunnelConfig.Password,
		PrivateKey:        sshTunnelConfig.PrivateKey,
		LocalForwardPort:  registry.Port,
		RemoteForwardPort: registry.Port,
	}
	sshTunnel := m.sshTunnelFactory.NewSSHTunnel(sshTunnelOptions)
	sshReadyCh := make(chan struct{})
	sshErrCh := make(chan error)
	go sshTunnel.Start(sshReadyCh, sshErrCh)
	defer sshTunnel.Stop()

	select {
	case <-sshReadyCh:
	case err := <-sshErrCh:
		return bosherr.WrapError(err, "Starting SSH tunnel")
	}

	return nil
}

func (m *microDeployer) startRegistry(registry bmdepl.Registry, readyCh chan struct{}) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
}
