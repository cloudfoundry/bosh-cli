package vm

import (
	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	biagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
	bihttpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type Manager interface {
	FindCurrent() (VM, bool, error)
	Create(bistemcell.CloudStemcell, bideplmanifest.Manifest) (VM, error)
}

type manager struct {
	vmRepo             biconfig.VMRepo
	stemcellRepo       biconfig.StemcellRepo
	diskDeployer       DiskDeployer
	agentClient        biagentclient.AgentClient
	agentClientFactory bihttpagent.AgentClientFactory
	cloud              bicloud.Cloud
	uuidGenerator      boshuuid.Generator
	fs                 boshsys.FileSystem
	logger             boshlog.Logger
	logTag             string
}

func NewManager(
	vmRepo biconfig.VMRepo,
	stemcellRepo biconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	agentClient biagentclient.AgentClient,
	cloud bicloud.Cloud,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Manager {
	return &manager{
		cloud:         cloud,
		agentClient:   agentClient,
		vmRepo:        vmRepo,
		stemcellRepo:  stemcellRepo,
		diskDeployer:  diskDeployer,
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
		logTag:        "vmManager",
	}
}

func (m *manager) FindCurrent() (VM, bool, error) {
	vmCID, found, err := m.vmRepo.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding currently deployed vm")
	}

	if !found {
		return nil, false, nil
	}

	vm := NewVM(
		vmCID,
		m.vmRepo,
		m.stemcellRepo,
		m.diskDeployer,
		m.agentClient,
		m.cloud,
		m.fs,
		m.logger,
	)

	return vm, true, err
}

func (m *manager) Create(stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest) (VM, error) {
	jobName := deploymentManifest.JobName()
	networkInterfaces, err := deploymentManifest.NetworkInterfaces(jobName)
	m.logger.Debug(m.logTag, "Creating VM with network interfaces: %#v", networkInterfaces)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting network spec")
	}

	resourcePool, err := deploymentManifest.ResourcePool(jobName)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting resource pool for job '%s'", jobName)
	}

	agentID, err := m.uuidGenerator.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating agent ID")
	}

	cid, err := m.cloud.CreateVM(agentID, stemcell.CID(), resourcePool.CloudProperties, networkInterfaces, resourcePool.Env)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating vm with stemcell cid '%s'", stemcell.CID())
	}

	metadata := bicloud.VMMetadata{
		Deployment: deploymentManifest.Name,
		Job:        deploymentManifest.JobName(),
		Index:      "0",
		Director:   "bosh-init",
	}
	err = m.cloud.SetVMMetadata(cid, metadata)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Setting VM metadata to %s", metadata)
	}

	err = m.vmRepo.UpdateCurrent(cid)
	if err != nil {
		return nil, bosherr.WrapError(err, "Updating current vm record")
	}

	vm := NewVM(
		cid,
		m.vmRepo,
		m.stemcellRepo,
		m.diskDeployer,
		m.agentClient,
		m.cloud,
		m.fs,
		m.logger,
	)

	return vm, nil
}
