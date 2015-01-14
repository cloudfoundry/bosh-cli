package set

import (
	version "github.com/hashicorp/go-version"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
)

type Resolver interface {
	Filter(releases []bmrelmanifest.ReleaseRef) error
	Find(name string) (release bmrel.Release, err error)
}

type resolver struct {
	manager bmrel.Manager
	logger  boshlog.Logger
	logTag  string

	releaseMap map[string]bmrelmanifest.ReleaseRef
}

func NewResolver(
	manager bmrel.Manager,
	logger boshlog.Logger,
) Resolver {
	return &resolver{
		manager: manager,
		logger:  logger,
		logTag:  "releaseResolver",
	}
}

func (r *resolver) Filter(releases []bmrelmanifest.ReleaseRef) error {
	r.releaseMap = map[string]bmrelmanifest.ReleaseRef{}
	for _, releaseRef := range releases {
		_, found := r.releaseMap[releaseRef.Name]
		if found {
			return bosherr.Errorf("Duplicate release '%s'", releaseRef.Name)
		}
		r.releaseMap[releaseRef.Name] = releaseRef
	}
	return nil
}

func (r *resolver) Find(name string) (release bmrel.Release, err error) {
	releases, found := r.manager.FindByName(name)
	if !found {
		return nil, bosherr.Errorf("Release '%s' is not available", name)
	}

	if r.releaseMap == nil || len(r.releaseMap) == 0 {
		//no releases specified, all provided releases are allowed, latest is used
		return r.findLatest(releases)
	}

	// if any releases are specified, all need to be specified
	releaseRef, releaseSpecified := r.releaseMap[name]
	if !releaseSpecified {
		return nil, bosherr.Errorf("Release '%s' was not specified", name)
	}

	if releaseRef.IsLatest() {
		return r.findLatest(releases)
	}

	return r.findByVersion(releases, releaseRef)
}

func (r *resolver) findLatest(releases []bmrel.Release) (bmrel.Release, error) {
	// assumes non-empty input array
	if len(releases) == 1 {
		return releases[0], nil
	}
	var (
		latest        bmrel.Release
		latestVersion *version.Version
	)
	for _, release := range releases {
		parsedVersion, err := version.NewVersion(release.Version())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Parsing version '%s' of release '%s", release.Version(), release.Name())
		}

		if latestVersion == nil || parsedVersion.GreaterThan(latestVersion) {
			latestVersion = parsedVersion
			latest = release
		}
	}

	return latest, nil
}

func (r *resolver) findByVersion(releases []bmrel.Release, releaseRef bmrelmanifest.ReleaseRef) (bmrel.Release, error) {
	// assumes non-empty input array

	versionConstraint, err := version.NewConstraint(releaseRef.Version)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing version '%s' of release '%s' from manifest", releaseRef.Version, releaseRef.Name)
	}

	for _, release := range releases {
		parsedVersion, err := version.NewVersion(release.Version())
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Parsing version '%s' of release '%s'", release.Version(), release.Name())
		}

		if versionConstraint.Check(parsedVersion) {
			return release, nil
		}
	}

	return nil, bosherr.Errorf("No version of '%s' matches '%s'", releaseRef.Name, releaseRef.Version)
}
