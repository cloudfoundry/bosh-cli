package stemcell

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type ManagerFactory interface {
	NewManager(Infrastructure) Manager
}

type managerFactory struct {
	fs     boshsys.FileSystem
	reader Reader
	repo   Repo
}

func NewManagerFactory(fs boshsys.FileSystem, reader Reader, repo Repo) ManagerFactory {
	return &managerFactory{
		fs:     fs,
		reader: reader,
		repo:   repo,
	}
}

func (f *managerFactory) NewManager(infrastructure Infrastructure) Manager {
	return NewManager(f.fs, f.reader, f.repo, infrastructure)
}
