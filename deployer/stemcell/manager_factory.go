package stemcell

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
)

type ManagerFactory interface {
	NewManager(Infrastructure) Manager
}

type managerFactory struct {
	fs          boshsys.FileSystem
	reader      Reader
	repo        Repo
	eventLogger bmeventlog.EventLogger
}

func NewManagerFactory(fs boshsys.FileSystem, reader Reader, repo Repo, eventLogger bmeventlog.EventLogger) ManagerFactory {
	return &managerFactory{
		fs:          fs,
		reader:      reader,
		repo:        repo,
		eventLogger: eventLogger,
	}
}

func (f *managerFactory) NewManager(infrastructure Infrastructure) Manager {
	return &manager{
		fs:             f.fs,
		reader:         f.reader,
		repo:           f.repo,
		eventLogger:    f.eventLogger,
		infrastructure: infrastructure,
	}
}
