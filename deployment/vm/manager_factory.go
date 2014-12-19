package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
)

type ManagerFactory interface {
	NewManager(cloud bmcloud.Cloud, mbusURL string) Manager
}

type managerFactory struct {
	vmRepo                 bmconfig.VMRepo
	stemcellRepo           bmconfig.StemcellRepo
	diskDeployer           DiskDeployer
	agentClientFactory     bmagentclient.Factory
	applySpecFactory       bmas.Factory
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
}

func NewManagerFactory(
	vmRepo bmconfig.VMRepo,
	stemcellRepo bmconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	agentClientFactory bmagentclient.Factory,
	applySpecFactory bmas.Factory,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		vmRepo:                 vmRepo,
		stemcellRepo:           stemcellRepo,
		diskDeployer:           diskDeployer,
		agentClientFactory:     agentClientFactory,
		applySpecFactory:       applySpecFactory,
		templatesSpecGenerator: templatesSpecGenerator,
		fs:     fs,
		logger: logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, mbusURL string) Manager {
	return NewManager(
		f.vmRepo,
		f.stemcellRepo,
		f.diskDeployer,
		mbusURL,
		f.agentClientFactory,
		f.templatesSpecGenerator,
		f.applySpecFactory,
		cloud,
		f.fs,
		f.logger,
	)
}
