package deployment

import (
	"net/http"

	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"

	biblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
)

type ManagerFactory interface {
	NewManager(cloud bicloud.Cloud, agentClientFactory bihttpagent.AgentClientFactory, directorID, mbusURL, caCert string, blobstoreFactory biblobstore.Factory, blobstoreHTTPClient *http.Client) Manager
}

type managerFactory struct {
	vmManagerFactory       bivm.ManagerFactory
	instanceManagerFactory biinstance.ManagerFactory
	diskManagerFactory     bidisk.ManagerFactory
	stemcellManagerFactory bistemcell.ManagerFactory
	deploymentFactory      Factory
}

func NewManagerFactory(
	vmManagerFactory bivm.ManagerFactory,
	instanceManagerFactory biinstance.ManagerFactory,
	diskManagerFactory bidisk.ManagerFactory,
	stemcellManagerFactory bistemcell.ManagerFactory,
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

func (f *managerFactory) NewManager(cloud bicloud.Cloud, agentClientFactory bihttpagent.AgentClientFactory, directorID, mbusURL, caCert string, blobstoreFactory biblobstore.Factory, blobstoreHTTPClient *http.Client) Manager {
	vmManager := f.vmManagerFactory.NewManager(cloud, agentClientFactory, directorID, mbusURL, caCert)
	instanceManager := f.instanceManagerFactory.NewManager(cloud, vmManager, blobstoreFactory, blobstoreHTTPClient)
	diskManager := f.diskManagerFactory.NewManager(cloud)
	stemcellManager := f.stemcellManagerFactory.NewManager(cloud)

	return NewManager(instanceManager, diskManager, stemcellManager, f.deploymentFactory)
}
