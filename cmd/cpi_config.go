package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type CPIConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCPIConfigCmd(ui boshui.UI, director boshdir.Director) CPIConfigCmd {
	return CPIConfigCmd{ui: ui, director: director}
}

func (c CPIConfigCmd) Run() error {
	cpiConfig, err := c.director.LatestCPIConfig()
	if err != nil {
		return err
	}

	c.ui.PrintBlock([]byte(cpiConfig.Properties))

	return nil
}
