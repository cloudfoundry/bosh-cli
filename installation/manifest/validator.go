package manifest

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
)

type Validator interface {
	Validate(Manifest, birelsetmanifest.Manifest) error
}

type validator struct {
	logger boshlog.Logger
}

func NewValidator(logger boshlog.Logger) Validator {
	return &validator{
		logger: logger,
	}
}

func (v *validator) Validate(manifest Manifest, releaseSetManifest birelsetmanifest.Manifest) error {
	var errs []error

	// When there is nothing in templates, return an error. It should have a CPI release.
	if len(manifest.Templates) == 0 {
		return fmt.Errorf("either cloud_provider.templates or cloud_provider.template must be provided and must contain at least one release")
	}

	for _, template := range manifest.Templates {
		errRet := v.validateReleaseJobRef(template, releaseSetManifest)
		errs = append(errs, errRet...)
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) validateReleaseJobRef(releaseJobRef ReleaseJobRef, releaseSetManifest birelsetmanifest.Manifest) []error {
	var errs []error
	jobName := releaseJobRef.Name
	if v.isBlank(jobName) {
		errs = append(errs, bosherr.Error("cloud_provider.template.name must be provided"))
	}

	releaseName := releaseJobRef.Release
	if v.isBlank(releaseName) {
		errs = append(errs, bosherr.Error("cloud_provider.template.release must be provided"))
	}

	_, found := releaseSetManifest.FindByName(releaseName)
	if !found {
		errs = append(errs, bosherr.Errorf("cloud_provider.template.release '%s' must refer to a release in releases", releaseName))
	}
	return errs
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
