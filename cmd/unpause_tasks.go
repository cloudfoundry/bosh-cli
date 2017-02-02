package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type UnpauseTasksCmd struct {
	director boshdir.Director
}

func NewUnpauseTasksCmd(director boshdir.Director) UnpauseTasksCmd {
	return UnpauseTasksCmd{director: director}
}

func (c UnpauseTasksCmd) Run() error {
	return c.director.PauseTasks(false)
}
