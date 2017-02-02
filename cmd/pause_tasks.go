package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type PauseTasksCmd struct {
	director boshdir.Director
}

func NewPauseTasksCmd(director boshdir.Director) PauseTasksCmd {
	return PauseTasksCmd{director: director}
}

func (c PauseTasksCmd) Run() error {
	return c.director.PauseTasks(true)
}
