package vm

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
)

type ManagerFactory interface {
	NewManager(infrastructure Infrastructure) Manager
}

type managerFactory struct {
	eventLogger bmeventlog.EventLogger
}

func NewManagerFactory(eventLogger bmeventlog.EventLogger) ManagerFactory {
	return &managerFactory{
		eventLogger: eventLogger,
	}
}

func (f *managerFactory) NewManager(infrastructure Infrastructure) Manager {
	return &manager{
		infrastructure: infrastructure,
		eventLogger:    f.eventLogger,
	}
}
