package tar

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type cmdExtractor struct {
	runner boshsys.CmdRunner
	logger boshlog.Logger
}

func NewCmdExtractor(runner boshsys.CmdRunner, logger boshlog.Logger) cmdExtractor {
	return cmdExtractor{
		runner: runner,
		logger: logger,
	}
}

func (e cmdExtractor) Extract(source string, destination string) error {
	_, _, _, err := e.runner.RunCommand("tar", "-C", destination, "-xzf", source)
	if err != nil {
		return bosherr.WrapError(err, "Extracting tar `%s' to `%s'", source, destination)
	}

	return nil
}
