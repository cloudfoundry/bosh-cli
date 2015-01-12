package manifest

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmerr "github.com/cloudfoundry/bosh-micro-cli/release/errors"
)

type Validator interface {
	Validate(Manifest) error
}

type validator struct {
	logger         boshlog.Logger
	releaseManager bmrel.Manager
}

func NewValidator(logger boshlog.Logger, releaseManager bmrel.Manager) Validator {
	return &validator{
		logger:         logger,
		releaseManager: releaseManager,
	}
}

func (v *validator) Validate(manifest Manifest) error {
	errs := []error{}
	releaseNames := map[string]struct{}{}
	for releaseIdx, release := range manifest.Releases {
		if v.isBlank(release.Name) {
			errs = append(errs, bosherr.Errorf("releases[%d].name must be provided", releaseIdx))
		}

		if v.isBlank(release.Version) {
			errs = append(errs, bosherr.Errorf("releases[%d].version must be provided", releaseIdx))
		}

		if !release.IsLatest() {
			if _, err := release.VersionConstraints(); err != nil {
				errs = append(errs, bosherr.WrapErrorf(err, "releases[%d].version must be a semantic version", releaseIdx))
			}
		}

		if _, found := releaseNames[release.Name]; found {
			errs = append(errs, bosherr.Errorf("releases[%d].name '%s' must be unique", releaseIdx, release.Name))
		}
		releaseNames[release.Name] = struct{}{}
	}

	releaseResolver := bmrel.NewResolver(v.logger, v.releaseManager, manifest.Releases)
	for releaseIdx, release := range manifest.Releases {
		_, err := releaseResolver.Find(release.Name)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err, "releases[%d] must refer to an available release", releaseIdx))
		}
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
