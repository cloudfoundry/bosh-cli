package release

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	birel "github.com/cloudfoundry/bosh-cli/v7/release"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

const (
	ReleaseBinaryName = "bin/cpi"
)

type CpiInstaller struct {
	ReleaseManager   birel.Manager
	InstallerFactory biinstall.InstallerFactory
}

func (i CpiInstaller) ValidateCpiRelease(installationManifest biinstallmanifest.Manifest, stage biui.Stage) error {
	return stage.Perform("Validating cpi release", func() error {
		var (
			errs                   []error
			releasePackagingErrs   []error
			releaseNamesInspected  []string
			numberCpiBinariesFound = 0
		)

		for _, template := range installationManifest.Templates {
			releaseName := template.Release
			releaseJobName := template.Name
			release, found := i.ReleaseManager.Find(releaseName)
			releaseNamesInspected = append(releaseNamesInspected, releaseName)

			if !found {
				releasePackagingErrs = append(releasePackagingErrs, bosherr.Errorf("installation release '%s' must refer to a provided release", releaseName))
				continue
			}

			job, ok := release.FindJobByName(releaseJobName)

			if !ok {
				releasePackagingErrs = append(releasePackagingErrs, bosherr.Errorf("release '%s' must contain specified job '%s'", releaseName, releaseJobName))
				continue
			}

			_, ok = job.FindTemplateByValue(ReleaseBinaryName)
			if ok {
				numberCpiBinariesFound += 1
			}
		}

		if numberCpiBinariesFound != 1 {
			errs = append(errs, bosherr.Errorf("Found %d releases containing a template that renders to target '%s'. Expected to find 1. Releases inspected: %v", numberCpiBinariesFound, ReleaseBinaryName, releaseNamesInspected))
			errs = append(errs, releasePackagingErrs...)
			return bosherr.NewMultiError(errs...)
		} else {
			return nil
		}
	})
}

func (i CpiInstaller) installCpiRelease(installer biinstall.Installer, installationManifest biinstallmanifest.Manifest, target biinstall.Target, stage biui.Stage) (biinstall.Installation, error) {
	var installation biinstall.Installation
	var err error
	err = stage.PerformComplex("installing CPI", func(installStage biui.Stage) error {
		installation, err = installer.Install(installationManifest, installStage)
		return err
	})
	if err != nil {
		return installation, bosherr.WrapError(err, "Installing CPI")
	}

	return installation, nil
}

func (i CpiInstaller) WithInstalledCpiRelease(installationManifest biinstallmanifest.Manifest, target biinstall.Target, stage biui.Stage, fn func(biinstall.Installation) error) (errToReturn error) {
	installer := i.InstallerFactory.NewInstaller(target)

	installation, err := i.installCpiRelease(installer, installationManifest, target, stage)
	if err != nil {
		errToReturn = err
		return
	}

	defer func() {
		err = i.cleanupInstall(installation, installer, stage)
		if errToReturn == nil {
			errToReturn = err
		}
	}()

	errToReturn = fn(installation)
	return
}

func (i CpiInstaller) cleanupInstall(installation biinstall.Installation, installer biinstall.Installer, stage biui.Stage) error {
	return stage.Perform("Cleaning up rendered CPI jobs", func() error {
		return installer.Cleanup(installation)
	})
}
