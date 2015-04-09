package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmrel "github.com/cloudfoundry/bosh-init/release"
)

const (
	ReleaseBinaryName = "bin/cpi"
)

type Validator struct {
}

func NewValidator() Validator {
	return Validator{}
}

func (v Validator) Validate(release bmrel.Release, cpiReleaseJobName string) error {
	job, ok := release.FindJobByName(cpiReleaseJobName)
	if !ok {
		return bosherr.Errorf("CPI release must contain specified job '%s'", cpiReleaseJobName)
	}

	_, ok = job.FindTemplateByValue(ReleaseBinaryName)
	if !ok {
		return bosherr.Errorf("Specified CPI release job '%s' must contain a template that renders to target '%s'", cpiReleaseJobName, ReleaseBinaryName)
	}

	return nil
}
