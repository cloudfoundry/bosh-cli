package cmd

import (
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Cmd interface {
	Name() string

	Meta() Meta

	Run(biui.Stage, []string) error
}

type Meta struct {
	Synopsis string
	Usage    string
	Env      map[string]MetaEnv
}

type MetaEnv struct {
	Example     string
	Description string
}
