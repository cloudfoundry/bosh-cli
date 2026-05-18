package vm

import (
	"code.cloudfoundry.org/clock"
	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
)

type ManagerFactory interface {
	NewManager(cloud bicloud.Cloud, agentClientFactory bihttpagent.AgentClientFactory, directorID, mbusURL, caCert string) Manager
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

func (f *managerFactory) NewManager(cloud bicloud.Cloud, agentClientFactory bihttpagent.AgentClientFactory, directorID, mbusURL, caCert string) Manager {
	return NewManager(
		f.vmRepo,
		f.stemcellRepo,
		f.diskDeployer,
		agentClientFactory,
		directorID,
		mbusURL,
		caCert,
		cloud,
		f.uuidGenerator,
		f.fs,
		f.logger,
		clock.NewClock(),
	)
}
