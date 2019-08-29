package cmd

import cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
import . "github.com/cloudfoundry/bosh-cli/cmd/opts"

type UnaliasEnvCmd struct {
	config cmdconf.Config
}

func NewUnaliasEnvCmd(config cmdconf.Config) *UnaliasEnvCmd {
	return &UnaliasEnvCmd{config: config}
}

func (c UnaliasEnvCmd) Run(opts UnaliasEnvOpts) error {
	updatedConfig, err := c.config.UnaliasEnvironment(opts.Args.Alias)
	if err != nil {
		return err
	}
	err = updatedConfig.Save()
	if err != nil {
		return err
	}

	return nil
}
