package validation

import (
	"errors"
	"fmt"
	"path"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmerr "github.com/cloudfoundry/bosh-micro-cli/errors"
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type boshValidator struct {
	fs boshsys.FileSystem
}

func NewBoshValidator(fs boshsys.FileSystem) Validator {
	return &boshValidator{fs: fs}
}

func (v *boshValidator) Validate(release bmrelease.Release) error {
	errs := []error{}

	err := v.validateReleaseName(release)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release name"))
	}

	err = v.validateReleaseVersion(release)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release version"))
	}

	err = v.validateReleaseJobs(release)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release jobs"))
	}

	err = v.validateReleasePackages(release)
	if err != nil {
		errs = append(errs, bosherr.WrapError(err, "Validating release packages"))
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v *boshValidator) validateReleaseName(release bmrelease.Release) error {
	if release.Name == "" {
		return errors.New("Release name is missing")
	}

	return nil
}

func (v *boshValidator) validateReleaseVersion(release bmrelease.Release) error {
	if release.Version == "" {
		return errors.New("Release version is missing")
	}

	return nil
}

func (v *boshValidator) validateReleaseJobs(release bmrelease.Release) error {
	errs := []error{}
	for _, job := range release.Jobs {
		if job.Name == "" {
			errs = append(errs, errors.New("Job name is missing"))
		}

		if job.Version == "" {
			errs = append(errs, fmt.Errorf("Job `%s' version is missing", job.Name))
		}

		if job.Fingerprint == "" {
			errs = append(errs, fmt.Errorf("Job `%s' fingerprint is missing", job.Name))
		}

		if job.Sha1 == "" {
			errs = append(errs, fmt.Errorf("Job `%s' sha1 is missing", job.Name))
		}

		monitPath := path.Join(job.ExtractedPath, "monit")
		if !v.fs.FileExists(monitPath) {
			errs = append(errs, fmt.Errorf("Job `%s' is missing monit file", job.Name))
		}

		for template := range job.Templates {
			templatePath := path.Join(job.ExtractedPath, "templates", template)
			if !v.fs.FileExists(templatePath) {
				errs = append(errs, fmt.Errorf("Job `%s' is missing template `%s'", job.Name, template))
			}
		}

		for _, pkgName := range job.PackageNames {
			found := false
			for _, releasePackage := range release.Packages {
				if releasePackage.Name == pkgName {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, fmt.Errorf("Job `%s' requires `%s' which is not in the release", job.Name, pkgName))
			}
		}
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}

func (v *boshValidator) validateReleasePackages(release bmrelease.Release) error {
	errs := []error{}
	for _, pkg := range release.Packages {
		if pkg.Name == "" {
			errs = append(errs, errors.New("Package name is missing"))
		}

		if pkg.Version == "" {
			errs = append(errs, fmt.Errorf("Package `%s' version is missing", pkg.Name))
		}

		if pkg.Fingerprint == "" {
			errs = append(errs, fmt.Errorf("Package `%s' fingerprint is missing", pkg.Name))
		}

		if pkg.Sha1 == "" {
			errs = append(errs, fmt.Errorf("Package `%s' sha1 is missing", pkg.Name))
		}
	}

	if len(errs) > 0 {
		return bmerr.NewExplainableError(errs)
	}

	return nil
}
