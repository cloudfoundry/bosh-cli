package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type RuntimeConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewRuntimeConfigCmd(ui boshui.UI, director boshdir.Director) RuntimeConfigCmd {
	return RuntimeConfigCmd{ui: ui, director: director}
}

func (c RuntimeConfigCmd) Run(opts RuntimeConfigOpts) error {
	runtimeConfig, err := c.director.LatestRuntimeConfig(opts.Name)
	if err != nil {
		return err
	}

	c.ui.PrintBlock([]byte(runtimeConfig.Properties))

	return nil
}
