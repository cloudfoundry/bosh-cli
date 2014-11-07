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
	NewManager(bmcloud.Cloud) Manager
}

type managerFactory struct {
	agentClientFactory      bmagentclient.Factory
	deploymentConfigService bmconfig.DeploymentConfigService
	applySpecFactory        bmas.Factory
	templatesSpecGenerator  bmas.TemplatesSpecGenerator
	fs                      boshsys.FileSystem
	logger                  boshlog.Logger
}

func NewManagerFactory(
	agentClientFactory bmagentclient.Factory,
	deploymentConfigService bmconfig.DeploymentConfigService,
	applySpecFactory bmas.Factory,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		agentClientFactory:      agentClientFactory,
		deploymentConfigService: deploymentConfigService,
		applySpecFactory:        applySpecFactory,
		templatesSpecGenerator:  templatesSpecGenerator,
		fs:     fs,
		logger: logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud) Manager {
	return &manager{
		cloud:                   cloud,
		agentClientFactory:      f.agentClientFactory,
		deploymentConfigService: f.deploymentConfigService,
		applySpecFactory:        f.applySpecFactory,
		templatesSpecGenerator:  f.templatesSpecGenerator,
		logger:                  f.logger,
		fs:                      f.fs,
		logTag:                  "vmManager",
	}
}
