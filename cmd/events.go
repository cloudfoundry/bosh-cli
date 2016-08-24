package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type EventsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewEventsCmd(ui boshui.UI, director boshdir.Director) EventsCmd {
	return EventsCmd{ui: ui, director: director}
}

func (c EventsCmd) Run(opts EventsOpts) error {
	filter := boshdir.EventsFilter{
		BeforeID:       opts.BeforeID,
		Before:         opts.Before,
		After:          opts.After,
		DeploymentName: opts.DeploymentName,
		TaskID:         opts.TaskID,
		Instance:       opts.Instance,
	}

	events, err := c.director.Events(filter)
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "events",
		Header:  []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task ID", "Deployment", "Instance", "Context"},
	}

	for _, e := range events {
		id := e.ID()

		if e.ParentID() != "" {
			id += " <- " + e.ParentID()
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(id),
			boshtbl.NewValueTime(e.Timestamp()),
			boshtbl.NewValueString(e.User()),
			boshtbl.NewValueString(e.Action()),
			boshtbl.NewValueString(e.ObjectType()),
			boshtbl.NewValueString(e.ObjectName()),
			boshtbl.NewValueString(e.TaskID()),
			boshtbl.NewValueString(e.DeploymentName()),
			boshtbl.NewValueString(e.Instance()),
			boshtbl.NewValueInterface(e.Context()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
