package vm

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type CID string

type Manager interface {
	CreateVM(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error)
}

type manager struct {
	infrastructure Infrastructure
	eventLogger    bmlog.EventLogger
}

func (m *manager) CreateVM(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (CID, error) {
	event := bmlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
		Task:  fmt.Sprintf("Creating VM from %s", stemcellCID),
		Index: 1,
		State: bmlog.Started,
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
		event = bmlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   1,
			Task:    fmt.Sprintf("Creating VM from %s", stemcellCID),
			Index:   1,
			State:   bmlog.Failed,
			Message: err.Error(),
		}
		m.eventLogger.AddEvent(event)
		return "", bosherr.WrapError(err, "creating vm with stemcell cid `%s'", stemcellCID)
	}

	event = bmlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
		Task:  fmt.Sprintf("Creating VM from %s", stemcellCID),
		Index: 1,
		State: bmlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return cid, err
}
