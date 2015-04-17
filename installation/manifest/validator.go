package manifest

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	birel "github.com/cloudfoundry/bosh-init/release"
)

type Validator interface {
	Validate(Manifest) error
}

type validator struct {
	logger         boshlog.Logger
	releaseManager birel.Manager
}

func NewValidator(logger boshlog.Logger, releaseManager birel.Manager) Validator {
	return &validator{
		logger:         logger,
		releaseManager: releaseManager,
	}
}

func (v *validator) Validate(manifest Manifest) error {
	errs := []error{}

	cpiJobName := manifest.Template.Name
	if v.isBlank(cpiJobName) {
		errs = append(errs, bosherr.Error("cloud_provider.template.name must be provided"))
	}

	cpiReleaseName := manifest.Template.Release
	if v.isBlank(cpiReleaseName) {
		errs = append(errs, bosherr.Error("cloud_provider.template.release must be provided"))
	}

	_, found := v.releaseManager.FindByName(cpiReleaseName)
	if !found {
		errs = append(errs, bosherr.Errorf("cloud_provider.template.release '%s' must refer to a release in releases", cpiReleaseName))
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
