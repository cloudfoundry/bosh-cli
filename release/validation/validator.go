package validation

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type validator struct {
	ui            bmui.UI
	cpiPath       string
	boshValidator ReleaseValidator
	cpiValidator  ReleaseValidator
}

type ReleaseValidator interface {
	Validate(release bmrel.Release) error
}

func NewValidator(boshValidator, cpiValidator ReleaseValidator, ui bmui.UI) ReleaseValidator {
	return &validator{
		ui:            ui,
		boshValidator: boshValidator,
		cpiValidator:  cpiValidator,
	}
}

func (v *validator) Validate(release bmrel.Release) error {
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
