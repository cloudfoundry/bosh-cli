package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type EventsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewEventsCmd(ui boshui.UI, director boshdir.Director) EventsCmd {
	return EventsCmd{ui: ui, director: director}
}

func (c EventsCmd) Run(opts EventsOpts) error {
	return nil
}
