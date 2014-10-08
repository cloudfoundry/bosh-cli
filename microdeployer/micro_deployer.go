package microdeployer

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type Deployer interface {
	Deploy(bmcloud.Cloud, bmdepl.Deployment, bmstemcell.CID) error
}

type microDeployer struct {
	vmManagerFactory bmvm.ManagerFactory
}

func NewMicroDeployer(vmManagerFactory bmvm.ManagerFactory) *microDeployer {
	return &microDeployer{
		vmManagerFactory: vmManagerFactory,
	}
}

func (m *microDeployer) Deploy(cpi bmcloud.Cloud, deployment bmdepl.Deployment, stemcellCID bmstemcell.CID) error {
	vmManager := m.vmManagerFactory.NewManager(cpi)
	_, err := vmManager.CreateVM(stemcellCID, deployment)

	if err != nil {
		return bosherr.WrapError(err, "Creating VM")
	}

	return nil
}
