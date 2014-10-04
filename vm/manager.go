package vm

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type CID string

type Manager interface {
	CreateVM(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error)
}

type manager struct {
	infrastructure Infrastructure
	eventLogger    bmeventlog.EventLogger
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

	networksSpec, err := deployment.NetworksSpec()
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
