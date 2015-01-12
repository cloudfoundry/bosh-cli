package release

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	"github.com/cloudfoundry/bosh-agent/errors"
	version "github.com/hashicorp/go-version"

	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
)

type Resolver interface {
	Find(name string) (release Release, err error)
}

type resolver struct {
	logger boshlog.Logger
	logTag string

	manager             Manager
	releaseVersions     []bmrelmanifest.ReleaseRef
	releaseMap          map[string]bmrelmanifest.ReleaseRef
	releaseMapPopulated bool
}

func NewResolver(
	logger boshlog.Logger,
	manager Manager,
	releaseVersions []bmrelmanifest.ReleaseRef,
) Resolver {
	return &resolver{
		logger:          logger,
		logTag:          "releaseResolver",
		manager:         manager,
		releaseVersions: releaseVersions,
	}
}

func (r *resolver) Find(name string) (release Release, err error) {
	releases, found := r.manager.FindByName(name)
	if found {
		var latestVersion *version.Version

		err = r.populateReleaseMap()
		if err != nil {
			return nil, err
		}

		releaseRef, foundRule := r.releaseMap[name]
		var versionConstraints version.Constraints
		if foundRule {
			versionConstraints, err = releaseRef.VersionConstraints()
			if err != nil {
				return nil, err
			}
		}

		for _, thisRelease := range releases {
			thisVersion, err := version.NewVersion(thisRelease.Version())
			if err != nil {
				return nil, errors.WrapErrorf(err, "Parsing version of '%s'", thisRelease.Name())
			}

			if foundRule && !versionConstraints.Check(thisVersion) {
				continue
			}

			if latestVersion == nil || thisVersion.GreaterThan(latestVersion) {
				release = thisRelease
				latestVersion = thisVersion
			}
		}

		if release == nil && foundRule {
			return nil, errors.Errorf("No version of '%s' matches '%s'", name, versionConstraints)
		}
	}

	if release == nil {
		return nil, errors.Errorf("Release '%s' is not available", name)
	}
	return release, nil
}

func (r *resolver) populateReleaseMap() error {
	if !r.releaseMapPopulated {
		r.releaseMap = map[string]bmrelmanifest.ReleaseRef{}
		for _, releaseRef := range r.releaseVersions {
			_, found := r.releaseMap[releaseRef.Name]
			if found {
				return errors.Errorf("Duplicate release '%s'", releaseRef.Name)
			}
			if releaseRef.IsLatest() {
				// use the newest version; equivalent to no rule at all
			} else {
				r.releaseMap[releaseRef.Name] = releaseRef
			}
		}

		r.releaseMapPopulated = true
	}
	return nil
}
