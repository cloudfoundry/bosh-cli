package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Manager interface {
	Add(Release)
	List() []Release
	FindByName(name string) (releases []Release, found bool)
	Find(name, version string) (release Release, found bool)
	DeleteAll() error
}

type manager struct {
	logger boshlog.Logger
	logTag string

	releases []Release
}

func NewManager(
	logger boshlog.Logger,
) Manager {
	return &manager{
		logger:   logger,
		logTag:   "releaseManager",
		releases: []Release{},
	}
}

func (m *manager) Add(release Release) {
	m.logger.Info(m.logTag, "Adding extracted release '%s-%s'", release.Name(), release.Version())
	m.releases = append(m.releases, release)
}

func (m *manager) List() []Release {
	return append([]Release(nil), m.releases...)
}

func (m *manager) FindByName(name string) (releases []Release, found bool) {
	releases = []Release{}
	found = false
	for _, release := range m.releases {
		if release.Name() == name {
			releases = append(releases, release)
			found = true
		}
	}
	return
}

func (m *manager) Find(name, version string) (release Release, found bool) {
	for _, release := range m.releases {
		if release.Name() == name && release.Version() == version {
			return release, true
		}
	}
	return nil, false
}

func (m *manager) DeleteAll() error {
	for _, release := range m.releases {
		deleteErr := release.Delete()
		if deleteErr != nil {
			return bosherr.Errorf("Failed to delete extracted release '%s': %s", release.Name(), deleteErr.Error())
		}
	}
	m.releases = []Release{}
	return nil
}
