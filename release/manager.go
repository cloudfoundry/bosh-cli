package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Manager interface {
	Extract(releaseTarballPath string) (Release, error)
	List() []Release
	Find(name string) (release Release, found bool)
	DeleteAll() error
}

type manager struct {
	fs        boshsys.FileSystem
	extractor boshcmd.Compressor
	validator Validator
	logger    boshlog.Logger
	logTag    string

	releases []Release
}

func NewManager(
	fs boshsys.FileSystem,
	extractor boshcmd.Compressor,
	validator Validator,
	logger boshlog.Logger,
) Manager {
	return &manager{
		fs:        fs,
		extractor: extractor,
		validator: validator,
		logger:    logger,
		logTag:    "releaseManager",
		releases:  []Release{},
	}
}

// Extract decompresses a release tarball into a temp directory (release.extractedPath),
// parses the release manifest, decompresses the packages and jobs, and validates the release.
// Use release.Delete() to clean up the temp directory.
func (m *manager) Extract(releaseTarballPath string) (Release, error) {
	extractedReleasePath, err := m.fs.TempDir("bosh-micro-release")
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating temp directory to extract release '%s'", releaseTarballPath)
	}

	m.logger.Info(m.logTag, "Extracting release tarball '%s' to '%s'", releaseTarballPath, extractedReleasePath)

	releaseReader := NewReader(releaseTarballPath, extractedReleasePath, m.fs, m.extractor)
	release, err := releaseReader.Read()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading release from '%s'", releaseTarballPath)
	}

	err = m.validator.Validate(release)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Validating release '%s-%s'", release.Name(), release.Version())
	}

	m.logger.Info(m.logTag, "Adding extracted release '%s-%s'", release.Name(), release.Version())
	m.releases = append(m.releases, release)

	return release, nil
}

func (m *manager) List() []Release {
	return append([]Release(nil), m.releases...)
}

func (m *manager) Find(name string) (release Release, found bool) {
	for _, release := range m.releases {
		if release.Name() == name {
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
