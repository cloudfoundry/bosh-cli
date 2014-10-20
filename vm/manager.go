package vm

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type CID string

func (c CID) String() string {
	return string(c)
}

type Manager interface {
	CreateVM(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error)
}

type manager struct {
	infrastructure          Infrastructure
	eventLogger             bmeventlog.EventLogger
	deploymentConfigService bmconfig.DeploymentConfigService
	logTag                  string
	logger                  boshlog.Logger
}

func (m *manager) CreateVM(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error) {
	event := bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
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

	cid, err := m.infrastructure.CreateVM(stemcellCID, cloudProperties, networksSpec, env)
	if err != nil {
		event = bmeventlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   1,
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
	deploymentConfig.VMCID = cid.String()

	err = m.deploymentConfigService.Save(deploymentConfig)
	if err != nil {
		return "", bosherr.WrapError(err, "Saving deployment config")
	}

	event = bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
		Task:  fmt.Sprintf("Creating VM from %s", stemcellCID),
		Index: 1,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return cid, err
}
