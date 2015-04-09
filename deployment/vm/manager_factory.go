package vm

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	biagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
)

type ManagerFactory interface {
	NewManager(cloud bicloud.Cloud, agentClient biagentclient.AgentClient) Manager
}

type managerFactory struct {
	vmRepo        biconfig.VMRepo
	stemcellRepo  biconfig.StemcellRepo
	diskDeployer  DiskDeployer
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
	logger        boshlog.Logger
}

func NewManagerFactory(
	vmRepo biconfig.VMRepo,
	stemcellRepo biconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		vmRepo:        vmRepo,
		stemcellRepo:  stemcellRepo,
		diskDeployer:  diskDeployer,
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
	}
}

func (f *managerFactory) NewManager(cloud bicloud.Cloud, agentClient biagentclient.AgentClient) Manager {
	return NewManager(
		f.vmRepo,
		f.stemcellRepo,
		f.diskDeployer,
		agentClient,
		cloud,
		f.uuidGenerator,
		f.fs,
		f.logger,
	)
}
