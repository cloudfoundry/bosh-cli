package release

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type Validator struct {
}

func NewValidator() Validator {
	return Validator{}
}

func (v Validator) Validate(release bmrel.Release) error {
	errs := v.validateCpiJob(release)
	if len(errs) > 0 {
		wrappedErrs := []error{}
		for _, err := range errs {
			wrappedErrs = append(wrappedErrs, bosherr.WrapError(err, "Validating CPI release"))
		}
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v Validator) validateCpiJob(release bmrel.Release) []error {
	errs := []error{}

	job, ok := release.FindJobByName(ReleaseJobName)
	if !ok {
		return append(errs, bosherr.Errorf("Job '%s' is missing from release", ReleaseJobName))
	}

	_, ok = job.FindTemplateByValue(ReleaseBinaryName)
	if !ok {
		errs = append(errs, bosherr.Errorf("Job '%s' is missing '%s' target", ReleaseJobName, ReleaseBinaryName))
	}

	return errs
}
