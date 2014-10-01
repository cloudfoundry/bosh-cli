package vm

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type CID string

type Manager interface {
	CreateVM(stemcellCID bmstemcell.CID) (CID, error)
}

type manager struct {
	infrastructure Infrastructure
	eventLogger    bmlog.EventLogger
}

func (m *manager) CreateVM(stemcellCID bmstemcell.CID) (CID, error) {
	event := bmlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
		Task:  "creating vm",
		Index: 1,
		State: bmlog.Started,
	}
	m.eventLogger.AddEvent(event)

	cid, err := m.infrastructure.CreateVM(stemcellCID)
	if err != nil {
		event = bmlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   1,
			Task:    "creating vm",
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
		Task:  "creating vm",
		Index: 1,
		State: bmlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return cid, err
}
