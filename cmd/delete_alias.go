package cmd

import cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"

type DeleteAliasCmd struct {
	config cmdconf.Config
}

func NewDeleteAliasCmd(config cmdconf.Config) *DeleteAliasCmd {
	return &DeleteAliasCmd{config: config}
}

func (c DeleteAliasCmd) Run(opts DeleteAliasOpts) error {
	updatedConfig, err := c.config.DeleteAlias(opts.Args.Alias)
	if err != nil {
		return err
	}
	err = updatedConfig.Save()
	if err != nil {
		return err
	}

	return nil
}
