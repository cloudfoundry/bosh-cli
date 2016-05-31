package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type LogOutCmd struct {
	target string
	config cmdconf.Config
	ui     biui.UI
}

func NewLogOutCmd(target string, config cmdconf.Config, ui biui.UI) LogOutCmd {
	return LogOutCmd{target: target, config: config, ui: ui}
}

func (c LogOutCmd) Run() error {
	updatedConfig := c.config.UnsetCredentials(c.target)

	err := updatedConfig.Save()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Logged out from '%s'", c.target)

	return nil
}
