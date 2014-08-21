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
	commands map[string]Cmd
}

func NewFactory(
	config bmconfig.Config,
	configService bmconfig.Service,
	filesystem boshsys.FileSystem,
	ui bmui.UI,
	extractor bmtar.Extractor,
	releaseValidator bmrelvalidation.ReleaseValidator,
	releaseCompiler bmcomp.ReleaseCompiler,
	logger boshlog.Logger,
) Factory {
	return &factory{
		commands: map[string]Cmd{
			"deployment": NewDeploymentCmd(ui, config, configService, filesystem),
			"deploy": NewDeployCmd(
				ui,
				config,
				filesystem,
				extractor,
				releaseValidator,
				releaseCompiler,
				logger,
			),
		},
	}
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name], nil
}
