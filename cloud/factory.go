package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bminstall "github.com/cloudfoundry/bosh-micro-cli/install"
)

const (
	cpiJobName = "cpi"
)

type Factory interface {
	NewCloud([]bminstall.InstalledJob) (Cloud, error)
}

type factory struct {
	fs             boshsys.FileSystem
	cmdRunner      boshsys.CmdRunner
	deploymentUUID string
	logger         boshlog.Logger
}

func NewFactory(fs boshsys.FileSystem, cmdRunner boshsys.CmdRunner, deploymentUUID string, logger boshlog.Logger) Factory {
	return &factory{
		fs:             fs,
		cmdRunner:      cmdRunner,
		deploymentUUID: deploymentUUID,
		logger:         logger,
	}
}

func (f *factory) NewCloud(jobs []bminstall.InstalledJob) (Cloud, error) {
	// for now, the installed job must be named "cpi"
	cpiJob, found := f.findCPIJob(jobs)
	if !found {
		return nil, bosherr.New("No `%s' release job found in the CPI deployment", cpiJobName)
	}

	return NewCloud(f.fs, f.cmdRunner, cpiJob.Path, f.deploymentUUID, f.logger), nil
}

func (f *factory) findCPIJob(jobs []bminstall.InstalledJob) (cpiJob bminstall.InstalledJob, found bool) {
	for _, job := range jobs {
		if job.Name == cpiJobName {
			return job, true
		}
	}
	return bminstall.InstalledJob{}, false
}
