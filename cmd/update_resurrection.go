package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type UpdateResurrectionCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewUpdateResurrectionCmd(ui boshui.UI, director boshdir.Director) UpdateResurrectionCmd {
	return UpdateResurrectionCmd{ui: ui, director: director}
}

func (c UpdateResurrectionCmd) Run(opts UpdateResurrectionOpts) error {
	c.ui.ErrorLinef("This command has been deprecated in favor of resurrection-config. See https://bosh.io/docs/cli-v2/#update-resurrection for more details.")
	return c.director.EnableResurrection(bool(opts.Args.Enabled))
}
