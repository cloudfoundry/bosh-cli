package cloud

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
)

type Factory interface {
	NewCloud(installation biinstall.Installation, directorID string, stemcellApiVersion int) (Cloud, error)
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

func (f *factory) NewCloud(installation biinstall.Installation, directorID string, stemcellApiVersion int) (Cloud, error) {
	numberCpiBinariesFound := 0
	foundCPI := CPI{}

	for _, cpiJob := range installation.Jobs() {
		target := installation.Target()
		cpi := CPI{
			JobPath:     cpiJob.Path,
			JobsDir:     target.JobsPath(),
			PackagesDir: target.PackagesPath(),
		}

		cmdPath := cpi.ExecutablePath()
		if f.fs.FileExists(cmdPath) {
			numberCpiBinariesFound += 1
			foundCPI = cpi
		}
	}

	if numberCpiBinariesFound != 1 {
		return nil, bosherr.Errorf("Found %d Jobs with a 'bin/cpi' binary. Expected 1.", numberCpiBinariesFound)
	}

	cpiCmdRunner := NewCPICmdRunner(f.cmdRunner, foundCPI, f.logger)
	return NewCloud(cpiCmdRunner, directorID, stemcellApiVersion, f.logger), nil
}
