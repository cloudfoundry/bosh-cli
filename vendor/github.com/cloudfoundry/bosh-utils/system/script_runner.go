package system

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type ScriptRunner interface {
	Run(script string) (string, string, error)
}

type concreteScriptRunner struct {
	scriptCommandFactory ScriptCommandFactory
	cmdRunner            CmdRunner
	fs                   FileSystem
	logTag               string
	logger               boshlog.Logger
}

func NewConcreteScriptRunner(
	scriptCommandFactory ScriptCommandFactory,
	cmdRunner CmdRunner,
	fs FileSystem,
	logger boshlog.Logger,
) ScriptRunner {
	return concreteScriptRunner{
		scriptCommandFactory: scriptCommandFactory,
		cmdRunner:            cmdRunner,
		fs:                   fs,
		logTag:               "concreteScriptRunner",
		logger:               logger,
	}
}

func (r concreteScriptRunner) Run(script string) (string, string, error) {
	file, err := r.fs.TempFile(r.logTag)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Creating tempfile")
	}
	defer r.fs.RemoveAll(file.Name())

	// need to close file before renaming
	err = file.Close()
	if err != nil {
		return "", "", bosherr.WrapError(err, "Closing tempfile")
	}

	err = r.fs.WriteFileString(file.Name(), script)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Writing to tempfile")
	}

	filePathToRun := file.Name() + r.scriptCommandFactory.Extension()

	err = r.fs.Rename(file.Name(), filePathToRun)
	if err != nil {
		return "", "", bosherr.WrapError(err, "Renaming tempfile")
	}
	defer r.fs.RemoveAll(filePathToRun)

	r.logger.DebugWithDetails(r.logTag, "Running script", script)

	command := r.scriptCommandFactory.New(filePathToRun)
	stdout, stderr, _, err := r.cmdRunner.RunComplexCommand(command)

	return stdout, stderr, err
}
