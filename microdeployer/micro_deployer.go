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
	registryReadyCh := make(chan struct{})
	registryErrCh := make(chan error)
	go m.startRegistry(registry, registryReadyCh, registryErrCh)
	defer m.registryServer.Stop()

	<-registryReadyCh

	deployDoneCh := make(chan struct{})
	deployErrCh := make(chan error)
	go m.runDeploy(cpi, deployment, stemcellCID, deployDoneCh, deployErrCh)

	select {
	case err := <-registryErrCh:
		return err
	case err := <-deployErrCh:
		return err
	case <-deployDoneCh:
		return nil
	}
}

func (m *microDeployer) startRegistry(registry bmdepl.Registry, readyCh chan struct{}, errCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyCh)
	if err != nil {
		errCh <- bosherr.WrapError(err, "Running registry")
	}
}

func (m *microDeployer) runDeploy(
	cpi bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcellCID bmstemcell.CID,
	doneCh chan struct{},
	errCh chan error,
) {
	vmManager := m.vmManagerFactory.NewManager(cpi)
	_, err := vmManager.CreateVM(stemcellCID, deployment)
	if err != nil {
		errCh <- bosherr.WrapError(err, "Creating VM")
		return
	}

	doneCh <- struct{}{}
}
