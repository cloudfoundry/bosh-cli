package cmd

import (
	bmui "github.com/cloudfoundry/bosh-init/ui"
)

type Cmd interface {
	Run(bmui.Stage, []string) error
	Name() string
}
