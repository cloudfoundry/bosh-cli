package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
)

type ManagerFactory interface {
	NewManager(cloud bmcloud.Cloud, agentClient bmac.AgentClient, mbusURL string) Manager
}

type managerFactory struct {
	vmRepo                 bmconfig.VMRepo
	stemcellRepo           bmconfig.StemcellRepo
	diskDeployer           DiskDeployer
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	uuidGenerator          boshuuid.Generator
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
}

func NewManagerFactory(
	vmRepo bmconfig.VMRepo,
	stemcellRepo bmconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		vmRepo:                 vmRepo,
		stemcellRepo:           stemcellRepo,
		diskDeployer:           diskDeployer,
		templatesSpecGenerator: templatesSpecGenerator,
		uuidGenerator:          uuidGenerator,
		fs:                     fs,
		logger:                 logger,
	}
}

func (f *managerFactory) NewManager(cloud bmcloud.Cloud, agentClient bmac.AgentClient, mbusURL string) Manager {
	return NewManager(
		f.vmRepo,
		f.stemcellRepo,
		f.diskDeployer,
		agentClient,
		mbusURL,
		f.templatesSpecGenerator,
		cloud,
		f.uuidGenerator,
		f.fs,
		f.logger,
	)
}
