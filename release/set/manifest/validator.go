package manifest

import (
	"strings"

	version "github.com/hashicorp/go-version"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrelset "github.com/cloudfoundry/bosh-init/release/set"
)

type Validator interface {
	Validate(Manifest) error
}

type validator struct {
	logger          boshlog.Logger
	releaseResolver bmrelset.Resolver
}

func NewValidator(logger boshlog.Logger, releaseResolver bmrelset.Resolver) Validator {
	return &validator{
		logger:          logger,
		releaseResolver: releaseResolver,
	}
}

func (v *validator) Validate(manifest Manifest) error {
	errs := []error{}
	releaseNames := map[string]struct{}{}
	for releaseIdx, release := range manifest.Releases {
		if v.isBlank(release.Name) {
			errs = append(errs, bosherr.Errorf("releases[%d].name must be provided", releaseIdx))
		}

		if _, found := releaseNames[release.Name]; found {
			errs = append(errs, bosherr.Errorf("releases[%d].name '%s' must be unique", releaseIdx, release.Name))
		}
		releaseNames[release.Name] = struct{}{}

		if !v.isBlank(release.Version) && !release.IsLatest() {
			if _, err := version.NewVersion(release.Version); err != nil {
				errs = append(errs, bosherr.WrapErrorf(err, "releases[%d].version '%s' must be a semantic version (name: '%s')", releaseIdx, release.Version, release.Name))
			}
		}
	}

	for releaseIdx, release := range manifest.Releases {
		_, err := v.releaseResolver.Find(release.Name)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err, "releases[%d] must refer to an available release", releaseIdx))
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
