package vm

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
)

type CID string

func (c CID) String() string {
	return string(c)
}

type Manager interface {
	Create(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error)
}

type manager struct {
	cloud                   bmcloud.Cloud
	eventLogger             bmeventlog.EventLogger
	deploymentConfigService bmconfig.DeploymentConfigService
	logTag                  string
	logger                  boshlog.Logger
}

func (m *manager) Create(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error) {
	event := bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 5,
		Task:  fmt.Sprintf("Creating VM from %s", stemcellCID),
		Index: 1,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

	microBoshJobName := deployment.Jobs[0].Name
	networksSpec, err := deployment.NetworksSpec(microBoshJobName)
	m.logger.Debug(m.logTag, "Creating VM with network spec: %#v", networksSpec)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM with stemcellCID `%s'", stemcellCID)
	}

	resourcePool := deployment.ResourcePools[0]
	cloudProperties, err := resourcePool.CloudProperties()
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM with stemcellCID `%s'", stemcellCID)
	}

	env, err := resourcePool.Env()
	if err != nil {
		return "", bosherr.WrapError(err, "Creating VM with stemcellCID `%s'", stemcellCID)
	}

	cid, err := m.cloud.CreateVM(stemcellCID.String(), cloudProperties, networksSpec, env)
	if err != nil {
		event = bmeventlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   5,
			Task:    fmt.Sprintf("Creating VM from %s", stemcellCID),
			Index:   1,
			State:   bmeventlog.Failed,
			Message: err.Error(),
		}
		m.eventLogger.AddEvent(event)
		return "", bosherr.WrapError(err, "creating vm with stemcell cid `%s'", stemcellCID)
	}

	deploymentConfig, err := m.deploymentConfigService.Load()
	if err != nil {
		return "", bosherr.WrapError(err, "Reading existing deployment config")
	}
	deploymentConfig.VMCID = cid

	err = m.deploymentConfigService.Save(deploymentConfig)
	if err != nil {
		return "", bosherr.WrapError(err, "Saving deployment config")
	}

	event = bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 5,
		Task:  fmt.Sprintf("Creating VM from %s", stemcellCID),
		Index: 1,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return CID(cid), err
}
