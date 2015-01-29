package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
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
	instanceStateBuilderFactory StateBuilderFactory
}

func NewFactory(
	instanceStateBuilderFactory StateBuilderFactory,
) Factory {
	return &factory{
		instanceStateBuilderFactory: instanceStateBuilderFactory,
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
	instanceStateBuilder := f.instanceStateBuilderFactory.NewStateBuilder(blobstore, vm.AgentClient())

	return NewInstance(
		jobName,
		id,
		vm,
		vmManager,
		sshTunnelFactory,
		instanceStateBuilder,
		logger,
	)
}
