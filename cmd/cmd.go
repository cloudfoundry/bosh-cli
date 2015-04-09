package cmd

import (
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Cmd interface {
	Run(biui.Stage, []string) error
	Name() string
}
