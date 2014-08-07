package release

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmerr "github.com/cloudfoundry/bosh-micro-cli/errors"
)

type Validator struct {
	fs      boshsys.FileSystem
	release Release
}

func NewValidator(fs boshsys.FileSystem, release Release) Validator {
	return Validator{
		fs:      fs,
		release: release,
	}
}

func (v Validator) Validate() error {
	errs := []error{}

	err := v.validateReleaseName()
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release name"))
	}

	err = v.validateReleaseVersion()
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release version"))
	}

	err = v.validateReleaseJobs()
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release jobs"))
	}

	err = v.validateReleasePackages()
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release packages"))
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v Validator) validateReleaseName() error {
	if v.release.Name == "" {
		return errors.New("Release name is missing")
	}

	return nil
}

func (v Validator) validateReleaseVersion() error {
	if v.release.Version == "" {
		return errors.New("Release version is missing")
	}

	return nil
}

func (v Validator) validateReleaseJobs() error {
	errs := []error{}
	for _, job := range v.release.Jobs {
		if job.Name == "" {
			errs = append(errs, errors.New("Job name is missing"))
		}

		if job.Version == "" {
			errs = append(errs, fmt.Errorf("Job '%s' version is missing", job.Name))
		}

		if job.Fingerprint == "" {
			errs = append(errs, fmt.Errorf("Job '%s' fingerprint is missing", job.Name))
		}

		if job.Sha1 == "" {
			errs = append(errs, fmt.Errorf("Job '%s' sha1 is missing", job.Name))
		}

		for template, templateFile := range job.Templates {
			if !v.fs.FileExists(templateFile) {
				errs = append(errs, fmt.Errorf("Job `%s' is missing template `%s'", job.Name, template))
			}
		}

		for _, pkg := range job.Packages {
			found := false
			for _, releasePackage := range v.release.Packages {
				if releasePackage.Name == pkg {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, fmt.Errorf("Job `%s' requires `%s' which is not in the release", job.Name, pkg))
			}
		}
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v Validator) validateReleasePackages() error {
	errs := []error{}
	for _, pkg := range v.release.Packages {
		if pkg.Name == "" {
			errs = append(errs, errors.New("Package name is missing"))
		}

		if pkg.Version == "" {
			errs = append(errs, fmt.Errorf("Package '%s' version is missing", pkg.Name))
		}

		if pkg.Fingerprint == "" {
			errs = append(errs, fmt.Errorf("Package '%s' fingerprint is missing", pkg.Name))
		}

		if pkg.Sha1 == "" {
			errs = append(errs, fmt.Errorf("Package '%s' sha1 is missing", pkg.Name))
		}
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}
