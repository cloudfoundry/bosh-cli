package vm

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type Manager interface {
	FindCurrent() (VM, bool, error)
	Create(bmstemcell.CloudStemcell, bmdepl.Deployment) (VM, error)
}

type manager struct {
	vmRepo                  bmconfig.VMRepo
	mbusURL                 string
	agentClientFactory      bmagentclient.Factory
	templatesSpecGenerator  bmas.TemplatesSpecGenerator
	deploymentConfigService bmconfig.DeploymentConfigService
	applySpecFactory        bmas.Factory
	cloud                   bmcloud.Cloud
	fs                      boshsys.FileSystem
	logger                  boshlog.Logger
	logTag                  string
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
		m.agentClient(),
		m.cloud,
		m.templatesSpecGenerator,
		m.applySpecFactory,
		m.mbusURL,
		m.fs,
		m.logger,
	)

	return vm, true, err
}

func (m *manager) Create(stemcell bmstemcell.CloudStemcell, deployment bmdepl.Deployment) (VM, error) {
	microBoshJobName := deployment.Jobs[0].Name
	networksSpec, err := deployment.NetworksSpec(microBoshJobName)
	m.logger.Debug(m.logTag, "Creating VM with network spec: %#v", networksSpec)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting network spec")
	}

	resourcePool := deployment.ResourcePools[0]
	cloudProperties, err := resourcePool.CloudProperties()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting cloud properties")
	}

	env, err := resourcePool.Env()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting resource pool env")
	}

	cid, err := m.cloud.CreateVM(stemcell.CID, cloudProperties, networksSpec, env)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating vm with stemcell cid `%s'", stemcell.CID)
	}

	deploymentConfig, err := m.deploymentConfigService.Load()
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading existing deployment config")
	}
	deploymentConfig.CurrentVMCID = cid

	err = m.deploymentConfigService.Save(deploymentConfig)
	if err != nil {
		return nil, bosherr.WrapError(err, "Saving deployment config")
	}

	vm := NewVM(
		cid,
		m.vmRepo,
		m.agentClient(),
		m.cloud,
		m.templatesSpecGenerator,
		m.applySpecFactory,
		m.mbusURL,
		m.fs,
		m.logger,
	)

	return vm, nil
}

func (m *manager) agentClient() bmagentclient.AgentClient {
	return m.agentClientFactory.Create(m.mbusURL)
}
