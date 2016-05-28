package system

import boshlog "github.com/cloudfoundry/bosh-utils/logger"

type ScriptRunner interface {
	Run(script string) (string, string, error)
}

type concreteScriptRunner struct {
	cmdRunner CmdRunner
	fs        FileSystem
	logTag    string
	logger    boshlog.Logger
}

func NewConcreteScriptRunner(
	cmdRunner CmdRunner,
	fs FileSystem,
	logger boshlog.Logger,
) ScriptRunner {
	return concreteScriptRunner{
		cmdRunner: cmdRunner,
		fs:        fs,
		logTag:    "concreteScriptRunner",
		logger:    logger,
	}
}
