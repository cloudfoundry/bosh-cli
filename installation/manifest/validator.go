package manifest

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
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
	cpiReleaseName := manifest.Release
	if v.isBlank(cpiReleaseName) {
		return bosherr.Error("cloud_provider.release must be provided")
	}

	_, err := v.releaseResolver.Find(cpiReleaseName)
	if err != nil {
		return bosherr.WrapErrorf(err, "cloud_provider.release '%s' must refer to a provided release", cpiReleaseName)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}
