package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bminstall "github.com/cloudfoundry/bosh-init/installation"
)

type Factory interface {
	NewCloud(installation bminstall.Installation, directorID string) (Cloud, error)
}

type factory struct {
	fs        boshsys.FileSystem
	cmdRunner boshsys.CmdRunner
	logger    boshlog.Logger
}

func NewFactory(
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	logger boshlog.Logger,
) Factory {
	return &factory{
		fs:        fs,
		cmdRunner: cmdRunner,
		logger:    logger,
	}
}

func (f *factory) NewCloud(installation bminstall.Installation, directorID string) (Cloud, error) {
	cpiJob := installation.Job()
	target := installation.Target()
	cpi := CPI{
		JobPath:     cpiJob.Path,
		JobsDir:     target.JobsPath(),
		PackagesDir: target.PackagesPath(),
	}

	cmdPath := cpi.ExecutablePath()
	if !f.fs.FileExists(cmdPath) {
		return nil, bosherr.Errorf("Installed CPI job '%s' does not contain the required executable '%s'", cpiJob.Name, cmdPath)
	}

	cpiCmdRunner := NewCPICmdRunner(f.cmdRunner, cpi, f.logger)
	return NewCloud(cpiCmdRunner, directorID, f.logger), nil
}
