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
	Create(bmstemcell.CloudStemcell, bmdepl.Deployment, string) (VM, error)
}

type manager struct {
	agentClientFactory      bmagentclient.Factory
	templatesSpecGenerator  bmas.TemplatesSpecGenerator
	deploymentConfigService bmconfig.DeploymentConfigService
	applySpecFactory        bmas.Factory
	cloud                   bmcloud.Cloud
	fs                      boshsys.FileSystem
	logger                  boshlog.Logger
	logTag                  string
}

func (m *manager) Create(stemcell bmstemcell.CloudStemcell, deployment bmdepl.Deployment, mbusURL string) (VM, error) {
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

	agentClient := m.agentClientFactory.Create(mbusURL)
	vm := NewVM(
		cid,
		agentClient,
		m.cloud,
		m.templatesSpecGenerator,
		m.applySpecFactory,
		mbusURL,
		m.fs,
		m.logger,
	)

	return vm, nil
}
