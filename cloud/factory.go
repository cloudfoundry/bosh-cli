package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
)

type Factory interface {
	NewCloud(installedCPIJob bmcpiinstall.InstalledJob, directorID string) (Cloud, error)
}

type factory struct {
	fs                  boshsys.FileSystem
	cmdRunner           boshsys.CmdRunner
	deploymentWorkspace bmconfig.DeploymentWorkspace
	logger              boshlog.Logger
}

func NewFactory(
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	deploymentWorkspace bmconfig.DeploymentWorkspace,
	logger boshlog.Logger,
) Factory {
	return &factory{
		fs:                  fs,
		cmdRunner:           cmdRunner,
		deploymentWorkspace: deploymentWorkspace,
		logger:              logger,
	}
}

func (f *factory) NewCloud(installedCPIJob bmcpiinstall.InstalledJob, directorID string) (Cloud, error) {
	cpiJob := CPIJob{
		JobPath:     installedCPIJob.Path,
		JobsDir:     f.deploymentWorkspace.JobsPath(),
		PackagesDir: f.deploymentWorkspace.PackagesPath(),
	}

	cmdPath := cpiJob.ExecutablePath()
	if !f.fs.FileExists(cmdPath) {
		return nil, bosherr.Errorf("Installed CPI job '%s' does not contain the required executable '%s'", installedCPIJob.Name, cmdPath)
	}

	cpiCmdRunner := NewCPICmdRunner(f.cmdRunner, cpiJob, f.logger)
	return NewCloud(cpiCmdRunner, directorID, f.logger), nil
}
