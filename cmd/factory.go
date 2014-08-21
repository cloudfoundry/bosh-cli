package cmd

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands         map[string](func() (Cmd, error))
	config           bmconfig.Config
	configService    bmconfig.Service
	fileSystem       boshsys.FileSystem
	ui               bmui.UI
	extractor        bmtar.Extractor
	releaseValidator bmrelvalidation.ReleaseValidator
	releaseCompiler  bmcomp.ReleaseCompiler
	logger           boshlog.Logger
}

func NewFactory(
	config bmconfig.Config,
	configService bmconfig.Service,
	fileSystem boshsys.FileSystem,
	ui bmui.UI,
	extractor bmtar.Extractor,
	releaseValidator bmrelvalidation.ReleaseValidator,
	releaseCompiler bmcomp.ReleaseCompiler,
	logger boshlog.Logger,
) Factory {
	f := &factory{
		config:           config,
		configService:    configService,
		fileSystem:       fileSystem,
		ui:               ui,
		extractor:        extractor,
		releaseValidator: releaseValidator,
		releaseCompiler:  releaseCompiler,
		logger:           logger,
	}
	f.commands = map[string](func() (Cmd, error)){
		"deployment": f.createDeploymentCmd,
		"deploy":     f.createDeployCmd,
	}
	return f
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name]()
}

func (f *factory) createDeploymentCmd() (Cmd, error) {
	return NewDeploymentCmd(f.ui, f.config, f.configService, f.fileSystem), nil
}

func (f *factory) createDeployCmd() (Cmd, error) {
	return NewDeployCmd(
		f.ui,
		f.config,
		f.fileSystem,
		f.extractor,
		f.releaseValidator,
		f.releaseCompiler,
		f.logger,
	), nil
}
