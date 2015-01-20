package instance

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
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
		blobstoreURL string,
		logger boshlog.Logger,
	) Instance
}

type factory struct {
	templatesSpecGenerator bmas.TemplatesSpecGenerator
}

func NewFactory(
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
) Factory {
	return &factory{
		templatesSpecGenerator: templatesSpecGenerator,
	}
}

func (f *factory) NewInstance(
	jobName string,
	id int,
	vm bmvm.VM,
	vmManager bmvm.Manager,
	sshTunnelFactory bmsshtunnel.Factory,
	blobstoreURL string,
	logger boshlog.Logger,
) Instance {
	return NewInstance(
		jobName,
		id,
		vm,
		vmManager,
		sshTunnelFactory,
		f.templatesSpecGenerator,
		blobstoreURL,
		logger,
	)
}
