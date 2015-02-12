package cmd

import (
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Cmd interface {
	Run(bmui.Stage, []string) error
	Name() string
}
