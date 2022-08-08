package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type CloudConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCloudConfigCmd(ui boshui.UI, director boshdir.Director) CloudConfigCmd {
	return CloudConfigCmd{ui: ui, director: director}
}

func (c CloudConfigCmd) Run(opts CloudConfigOpts) error {
	cloudConfig, err := c.director.LatestCloudConfig(opts.Name)
	if err != nil {
		return err
	}

	c.ui.PrintBlock([]byte(cloudConfig.Properties))

	return nil
}
