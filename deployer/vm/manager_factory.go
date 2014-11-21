package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
)

type ManagerFactory interface {
	NewManager(cloud bmcloud.Cloud, mbusURL string) Manager
}

type managerFactory struct {
	vmRepo                 bmconfig.VMRepo
	agentClientFactory     bmagentclient.Factory
	applySpecFactory       bmas.Factory
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
}

func NewManagerFactory(
	vmRepo bmconfig.VMRepo,
	agentClientFactory bmagentclient.Factory,
	applySpecFactory bmas.Factory,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		vmRepo:                 vmRepo,
		agentClientFactory:     agentClientFactory,
		applySpecFactory:       applySpecFactory,
		templatesSpecGenerator: templatesSpecGenerator,
		fs:     fs,
		logger: logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, mbusURL string) Manager {
	return &manager{
		cloud:                  cloud,
		mbusURL:                mbusURL,
		vmRepo:                 f.vmRepo,
		agentClientFactory:     f.agentClientFactory,
		applySpecFactory:       f.applySpecFactory,
		templatesSpecGenerator: f.templatesSpecGenerator,
		logger:                 f.logger,
		fs:                     f.fs,
		logTag:                 "vmManager",
	}
}
