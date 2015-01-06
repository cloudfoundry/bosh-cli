package vm

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type Manager interface {
	FindCurrent() (VM, bool, error)
	Create(bmstemcell.CloudStemcell, bmdeplmanifest.Manifest) (VM, error)
}

type manager struct {
	vmRepo                 bmconfig.VMRepo
	stemcellRepo           bmconfig.StemcellRepo
	diskDeployer           DiskDeployer
	agentClient            bmac.AgentClient
	mbusURL                string
	agentClientFactory     bmhttpagent.AgentClientFactory
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	applySpecFactory       bmas.Factory
	cloud                  bmcloud.Cloud
	uuidGenerator          boshuuid.Generator
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
	logTag                 string
}

func NewManager(
	vmRepo bmconfig.VMRepo,
	stemcellRepo bmconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	agentClient bmac.AgentClient,
	mbusURL string,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	applySpecFactory bmas.Factory,
	cloud bmcloud.Cloud,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Manager {
	return &manager{
		cloud:                  cloud,
		agentClient:            agentClient,
		mbusURL:                mbusURL,
		vmRepo:                 vmRepo,
		stemcellRepo:           stemcellRepo,
		diskDeployer:           diskDeployer,
		applySpecFactory:       applySpecFactory,
		templatesSpecGenerator: templatesSpecGenerator,
		uuidGenerator:          uuidGenerator,
		fs:                     fs,
		logger:                 logger,
		logTag:                 "vmManager",
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
		m.templatesSpecGenerator,
		m.applySpecFactory,
		m.mbusURL,
		m.fs,
		m.logger,
	)

	return vm, true, err
}

func (m *manager) Create(stemcell bmstemcell.CloudStemcell, deploymentManifest bmdeplmanifest.Manifest) (VM, error) {
	microBoshJobName := deploymentManifest.Jobs[0].Name
	networksSpec, err := deploymentManifest.NetworksSpec(microBoshJobName)
	m.logger.Debug(m.logTag, "Creating VM with network spec: %#v", networksSpec)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting network spec")
	}

	resourcePool := deploymentManifest.ResourcePools[0]
	cloudProperties, err := resourcePool.CloudProperties()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting cloud properties")
	}

	env, err := resourcePool.Env()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting resource pool env")
	}

	agentID, err := m.uuidGenerator.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating agent ID")
	}

	cid, err := m.cloud.CreateVM(agentID, stemcell.CID(), cloudProperties, networksSpec, env)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating vm with stemcell cid '%s'", stemcell.CID())
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
		m.templatesSpecGenerator,
		m.applySpecFactory,
		m.mbusURL,
		m.fs,
		m.logger,
	)

	return vm, nil
}
