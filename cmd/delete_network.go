package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeleteNetworkCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDeleteNetworkCmd(ui boshui.UI, director boshdir.Director) DeleteNetworkCmd {
	return DeleteNetworkCmd{ui: ui, director: director}
}

func (c DeleteNetworkCmd) Run(opts DeleteNetworkOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	network, err := c.director.FindOrphanNetwork(opts.Args.Name)
	if err != nil {
		return err
	}

	return network.Delete()
}
