package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
)

const (
	cpiJobName = "cpi"
)

type Factory interface {
	NewCloud([]bmcpiinstall.InstalledJob) (Cloud, error)
}

type factory struct {
	fs        boshsys.FileSystem
	cmdRunner boshsys.CmdRunner
	config    bmconfig.DeploymentConfig
	logger    boshlog.Logger
}

func NewFactory(fs boshsys.FileSystem, cmdRunner boshsys.CmdRunner, config bmconfig.DeploymentConfig, logger boshlog.Logger) Factory {
	return &factory{
		fs:        fs,
		cmdRunner: cmdRunner,
		config:    config,
		logger:    logger,
	}
}

func (f *factory) NewCloud(jobs []bmcpiinstall.InstalledJob) (Cloud, error) {
	// for now, the installed job must be named "cpi"
	installedCPIJob, found := f.findCPIJob(jobs)
	if !found {
		return nil, bosherr.New("No `%s' release job found in the CPI deployment", cpiJobName)
	}

	cpiJob := CPIJob{
		JobPath:     installedCPIJob.Path,
		JobsDir:     f.config.JobsPath(),
		PackagesDir: f.config.PackagesPath(),
	}

	cpiCmdRunner := NewCPICmdRunner(f.cmdRunner, cpiJob, f.config.DeploymentUUID, f.logger)
	return NewCloud(cpiCmdRunner, f.config.DeploymentUUID, f.logger), nil
}

func (f *factory) findCPIJob(jobs []bmcpiinstall.InstalledJob) (cpiJob bmcpiinstall.InstalledJob, found bool) {
	for _, job := range jobs {
		if job.Name == cpiJobName {
			return job, true
		}
	}
	return cpiJob, false
}
