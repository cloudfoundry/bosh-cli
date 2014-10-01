package vm

import (
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
)

type ManagerFactory interface {
	NewManager(infrastructure Infrastructure) Manager
}

type managerFactory struct {
	eventLogger bmlog.EventLogger
}

func NewManagerFactory(eventLogger bmlog.EventLogger) ManagerFactory {
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
