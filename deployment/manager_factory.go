package deployment

import (
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type ManagerFactory interface {
	NewManager(bmcloud.Cloud, bmac.AgentClient, bmblobstore.Blobstore) Manager
}

type managerFactory struct {
	vmManagerFactory       bmvm.ManagerFactory
	instanceManagerFactory bminstance.ManagerFactory
	diskManagerFactory     bmdisk.ManagerFactory
	stemcellManagerFactory bmstemcell.ManagerFactory
	deploymentFactory      Factory
}

func NewManagerFactory(
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	diskManagerFactory bmdisk.ManagerFactory,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	deploymentFactory Factory,
) ManagerFactory {
	return &managerFactory{
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		diskManagerFactory:     diskManagerFactory,
		stemcellManagerFactory: stemcellManagerFactory,
		deploymentFactory:      deploymentFactory,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, agentClient bmac.AgentClient, blobstore bmblobstore.Blobstore) Manager {
	vmManager := f.vmManagerFactory.NewManager(cloud, agentClient)
	instanceManager := f.instanceManagerFactory.NewManager(cloud, vmManager, blobstore)
	diskManager := f.diskManagerFactory.NewManager(cloud)
	stemcellManager := f.stemcellManagerFactory.NewManager(cloud)

	return NewManager(instanceManager, diskManager, stemcellManager, f.deploymentFactory)
}
