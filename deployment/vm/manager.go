package vm

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type Manager interface {
	FindCurrent() (VM, bool, error)
	Create(bmstemcell.CloudStemcell, bmmanifest.Manifest) (VM, error)
}

type manager struct {
	vmRepo                 bmconfig.VMRepo
	stemcellRepo           bmconfig.StemcellRepo
	mbusURL                string
	agentClientFactory     bmagentclient.Factory
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	applySpecFactory       bmas.Factory
	cloud                  bmcloud.Cloud
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
	logTag                 string
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

func (m *manager) Create(stemcell bmstemcell.CloudStemcell, deploymentManifest bmmanifest.Manifest) (VM, error) {
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

	cid, err := m.cloud.CreateVM(stemcell.CID(), cloudProperties, networksSpec, env)
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
