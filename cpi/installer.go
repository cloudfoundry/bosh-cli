package cpi

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmcomp "github.com/cloudfoundry/bosh-micro-cli/cpi/compile"
	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmcpirel "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Installer interface {
	Install(manifest bmmanifest.CPIDeploymentManifest, directorID string) (bmcloud.Cloud, error)
}

type cpiInstaller struct {
	ui              bmui.UI
	fs              boshsys.FileSystem
	extractor       boshcmd.Compressor
	releaseManager  bmrel.Manager
	releaseCompiler bmcomp.ReleaseCompiler
	jobInstaller    bmcpiinstall.JobInstaller
	cloudFactory    bmcloud.Factory
	logger          boshlog.Logger
	logTag          string
}

func NewInstaller(
	ui bmui.UI,
	fs boshsys.FileSystem,
	extractor boshcmd.Compressor,
	releaseManager bmrel.Manager,
	releaseCompiler bmcomp.ReleaseCompiler,
	jobInstaller bmcpiinstall.JobInstaller,
	cloudFactory bmcloud.Factory,
	logger boshlog.Logger,
) Installer {
	return &cpiInstaller{
		ui:              ui,
		fs:              fs,
		extractor:       extractor,
		releaseManager:  releaseManager,
		releaseCompiler: releaseCompiler,
		jobInstaller:    jobInstaller,
		cloudFactory:    cloudFactory,
		logger:          logger,
		logTag:          "cpiInstaller",
	}
}

func (c *cpiInstaller) Install(manifest bmmanifest.CPIDeploymentManifest, directorID string) (bmcloud.Cloud, error) {
	c.logger.Info(c.logTag, "Installing CPI deployment '%s'", manifest.Name)
	c.logger.Debug(c.logTag, "Installing CPI deployment '%s' with manifest: %#v", manifest.Name, manifest)

	releaseRef := manifest.Release
	release, found := c.releaseManager.Find(releaseRef.Name, releaseRef.Version)
	if !found {
		c.ui.Error(fmt.Sprintf("Could not find CPI release '%s/%s'", releaseRef.Name, releaseRef.Version))
		return nil, bosherr.Errorf("CPI release '%s/%s' not found", releaseRef.Name, releaseRef.Version)
	}

	if !release.Exists() {
		c.ui.Error("Could not find extracted CPI release")
		return nil, bosherr.Errorf("Extracted CPI release does not exist")
	}

	err := c.releaseCompiler.Compile(release, manifest)
	if err != nil {
		c.ui.Error("Could not compile CPI release")
		return nil, bosherr.WrapError(err, "Compiling CPI release")
	}

	installedJob, err := c.installCPIJob(release)
	if err != nil {
		c.ui.Error("Could not install CPI deployment")
		return nil, bosherr.WrapError(err, "Installing CPI deployment job")
	}

	cloud, err := c.cloudFactory.NewCloud(installedJob, directorID)
	if err != nil {
		c.ui.Error("Invalid CPI deployment")
		return nil, bosherr.WrapError(err, "Validating CPI deployment job installation")
	}

	return cloud, nil
}

func (c *cpiInstaller) installCPIJob(release bmrel.Release) (bmcpiinstall.InstalledJob, error) {
	cpiJobName := bmcpirel.ReleaseJobName
	releaseJob, found := release.FindJobByName(cpiJobName)

	if !found {
		c.ui.Error(fmt.Sprintf("Could not find CPI job '%s' in release '%s'", cpiJobName, release.Name()))
		return bmcpiinstall.InstalledJob{}, bosherr.Errorf("Invalid CPI release: job '%s' not found in release '%s'", cpiJobName, release.Name())
	}

	installedJob, err := c.jobInstaller.Install(releaseJob)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Could not install job '%s'", releaseJob.Name))
		return bmcpiinstall.InstalledJob{}, bosherr.WrapErrorf(err, "Installing job '%s' for CPI release", releaseJob.Name)
	}

	return installedJob, nil
}
