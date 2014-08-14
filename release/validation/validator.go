package validation

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type validator struct {
	ui            bmui.UI
	cpiPath       string
	boshValidator Validator
	cpiValidator  Validator
}

type Validator interface {
	Validate(release bmrelease.Release) error
}

func NewValidator(boshValidator, cpiValidator Validator, ui bmui.UI) Validator {
	return &validator{
		ui:            ui,
		boshValidator: boshValidator,
		cpiValidator:  cpiValidator,
	}
}

func (v *validator) Validate(release bmrelease.Release) error {
	err := v.boshValidator.Validate(release)
	if err != nil {
		v.ui.Error(fmt.Sprintf("CPI release `%s' is not a valid BOSH release", release.TarballPath))
		return bosherr.WrapError(err, "Validating CPI release")
	}

	err = v.cpiValidator.Validate(release)
	if err != nil {
		v.ui.Error(fmt.Sprintf("CPI release `%s' is not a valid CPI release", release.TarballPath))
		return bosherr.WrapError(err, "Validating CPI release")
	}

	return nil
}
