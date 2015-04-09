package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bminstancestate "github.com/cloudfoundry/bosh-init/deployment/instance/state"
	bmsshtunnel "github.com/cloudfoundry/bosh-init/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-init/deployment/vm"
)

type Factory interface {
	NewInstance(
		jobName string,
		id int,
		vm bmvm.VM,
		vmManager bmvm.Manager,
		sshTunnelFactory bmsshtunnel.Factory,
		blobstore bmblobstore.Blobstore,
		logger boshlog.Logger,
	) Instance
}

type factory struct {
	stateBuilderFactory bminstancestate.BuilderFactory
}

func NewFactory(
	stateBuilderFactory bminstancestate.BuilderFactory,
) Factory {
	return &factory{
		stateBuilderFactory: stateBuilderFactory,
	}
}

func (f *factory) NewInstance(
	jobName string,
	id int,
	vm bmvm.VM,
	vmManager bmvm.Manager,
	sshTunnelFactory bmsshtunnel.Factory,
	blobstore bmblobstore.Blobstore,
	logger boshlog.Logger,
) Instance {
	stateBuilder := f.stateBuilderFactory.NewBuilder(blobstore, vm.AgentClient())

	return NewInstance(
		jobName,
		id,
		vm,
		vmManager,
		sshTunnelFactory,
		stateBuilder,
		logger,
	)
}
