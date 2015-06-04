package release

import (
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	birel "github.com/cloudfoundry/bosh-init/release"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type CpiInstaller struct {
	ReleaseManager birel.Manager
	Installer      biinstall.Installer
	Validator      Validator
}

func (i CpiInstaller) ValidateCpiRelease(installationManifest biinstallmanifest.Manifest, stage biui.Stage) error {
	return stage.Perform("Validating cpi release", func() error {
		cpiReleaseName := installationManifest.Template.Release
		cpiRelease, found := i.ReleaseManager.Find(cpiReleaseName)
		if !found {
			return bosherr.Errorf("installation release '%s' must refer to a provided release", cpiReleaseName)
		}

		err := i.Validator.Validate(cpiRelease, installationManifest.Template.Name)
		if err != nil {
			return bosherr.WrapErrorf(err, "Invalid CPI release '%s'", cpiReleaseName)
		}
		return nil
	})
}

func (i CpiInstaller) installCpiRelease(installationManifest biinstallmanifest.Manifest, stage biui.Stage) (biinstall.Installation, error) {
	var installation biinstall.Installation
	var err error
	err = stage.PerformComplex("installing CPI", func(installStage biui.Stage) error {
		installation, err = i.Installer.InstallPackagesAndJobs(installationManifest, installStage)
		return err
	})
	if err != nil {
		return installation, bosherr.WrapError(err, "Installing CPI")
	}

	return installation, nil
}

func (i CpiInstaller) WithInstalledCpiRelease(installationManifest biinstallmanifest.Manifest, stage biui.Stage, fn func(biinstall.Installation) error) (errToReturn error) {
	installation, err := i.installCpiRelease(installationManifest, stage)
	if err != nil {
		errToReturn = err
		return
	}

	defer func() {
		err = i.cleanupInstall(installation, stage)
		if errToReturn == nil {
			errToReturn = err
		}
	}()

	errToReturn = fn(installation)
	return
}

func (i CpiInstaller) cleanupInstall(installation biinstall.Installation, stage biui.Stage) error {
	return stage.Perform("Cleaning up rendered CPI jobs", func() error {
		return i.Installer.Cleanup(installation)
	})
}
