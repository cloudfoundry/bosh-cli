package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type TaskConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewTaskConfigCmd(ui boshui.UI, director boshdir.Director) TaskConfigCmd {
	return TaskConfigCmd{ui: ui, director: director}
}

func (c TaskConfigCmd) Run() error {
	taskConfig, err := c.director.LatestTaskConfig()
	if err != nil {
		return err
	}

	c.ui.PrintBlock(taskConfig.Properties)

	return nil
}
