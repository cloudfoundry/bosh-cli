package deployment

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type ManagerFactory interface {
	NewManager(cloud bmcloud.Cloud, agentClient bmac.AgentClient, blobstoreURL string) Manager
}

type managerFactory struct {
	vmManagerFactory       bmvm.ManagerFactory
	instanceManagerFactory bminstance.ManagerFactory
	diskManagerFactory     bmdisk.ManagerFactory
	stemcellManagerFactory bmstemcell.ManagerFactory
}

func NewManagerFactory(
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	diskManagerFactory bmdisk.ManagerFactory,
	stemcellManagerFactory bmstemcell.ManagerFactory,
) ManagerFactory {
	return &managerFactory{
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		diskManagerFactory:     diskManagerFactory,
		stemcellManagerFactory: stemcellManagerFactory,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, agentClient bmac.AgentClient, blobstoreURL string) Manager {
	vmManager := f.vmManagerFactory.NewManager(cloud, agentClient, blobstoreURL)
	instanceManager := f.instanceManagerFactory.NewManager(cloud, vmManager)
	diskManager := f.diskManagerFactory.NewManager(cloud)
	stemcellManager := f.stemcellManagerFactory.NewManager(cloud)

	return NewManager(instanceManager, diskManager, stemcellManager)
}
