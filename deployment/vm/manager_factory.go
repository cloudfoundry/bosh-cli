package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
)

type ManagerFactory interface {
	NewManager(cloud bmcloud.Cloud, mbusURL string) Manager
}

type managerFactory struct {
	vmRepo                 bmconfig.VMRepo
	stemcellRepo           bmconfig.StemcellRepo
	diskDeployer           DiskDeployer
	agentClientFactory     bmhttpagent.AgentClientFactory
	applySpecFactory       bmas.Factory
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	uuidGenerator          boshuuid.Generator
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
}

func NewManagerFactory(
	vmRepo bmconfig.VMRepo,
	stemcellRepo bmconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	agentClientFactory bmhttpagent.AgentClientFactory,
	applySpecFactory bmas.Factory,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	uuidGenerator boshuuid.Generator,
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
		uuidGenerator:          uuidGenerator,
		fs:                     fs,
		logger:                 logger,
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
		f.uuidGenerator,
		f.fs,
		f.logger,
	)
}
