package microdeployer

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type Deployer interface {
	Deploy(bmcloud.Cloud, bmdepl.Deployment, bmstemcell.CID) error
}

type microDeployer struct {
	vmManagerFactory bmvm.ManagerFactory
	registryServer   bmregistry.Server
}

func NewMicroDeployer(vmManagerFactory bmvm.ManagerFactory, registryServer bmregistry.Server) *microDeployer {
	return &microDeployer{
		vmManagerFactory: vmManagerFactory,
		registryServer:   registryServer,
	}
}

func (m *microDeployer) Deploy(cpi bmcloud.Cloud, deployment bmdepl.Deployment, stemcellCID bmstemcell.CID) error {
	registry := deployment.Registry
	readyCh := make(chan struct{})
	go m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyCh)
	defer m.registryServer.Stop()

	<-readyCh

	vmManager := m.vmManagerFactory.NewManager(cpi)
	_, err := vmManager.CreateVM(stemcellCID, deployment)
	if err != nil {
		return bosherr.WrapError(err, "Creating VM")
	}

	return nil
}
