package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/cpideployer/install"
)

const (
	cpiJobName = "cpi"
)

type Factory interface {
	NewCloud([]bminstall.InstalledJob) (Cloud, error)
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

func (f *factory) NewCloud(jobs []bminstall.InstalledJob) (Cloud, error) {
	// for now, the installed job must be named "cpi"
	cpiJob, found := f.findCPIJob(jobs)
	if !found {
		return nil, bosherr.New("No `%s' release job found in the CPI deployment", cpiJobName)
	}

	cpi := CPIJob{
		JobPath:      cpiJob.Path,
		JobsPath:     f.config.JobsPath(),
		PackagesPath: f.config.PackagesPath(),
	}

	return NewCloud(f.fs, f.cmdRunner, cpi, f.config.DeploymentUUID, f.logger), nil
}

func (f *factory) findCPIJob(jobs []bminstall.InstalledJob) (cpiJob bminstall.InstalledJob, found bool) {
	for _, job := range jobs {
		if job.Name == cpiJobName {
			return job, true
		}
	}
	return cpiJob, false
}
