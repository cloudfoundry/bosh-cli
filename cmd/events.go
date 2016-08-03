package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
	//"fmt"
	//"encoding/json"
	"fmt"
)

type EventsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewEventsCmd(ui boshui.UI, director boshdir.Director) EventsCmd {
	return EventsCmd{ui: ui, director: director}
}

func (c EventsCmd) Run(opts EventsOpts) error {
	directorOpts := make(map[string]interface{})
	directorOpts["beforeId"] = opts.BeforeId
	directorOpts["before"] = opts.Before
	directorOpts["after"] = opts.After
	directorOpts["deployment"] = opts.Deployment
	directorOpts["task"] = opts.Task
	directorOpts["instance"] = opts.Instance
	return c.printTable(c.director.Events(directorOpts))
}

func (c EventsCmd) printTable(events []boshdir.Event, err error) error {
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "events",
		Header:  []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0}},
	}

	for _, e := range events {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueInt(e.Id()),
			boshtbl.NewValueTime(e.Timestamp()),
			boshtbl.NewValueString(e.User()),
			boshtbl.NewValueString(e.Action()),
			boshtbl.NewValueString(e.ObjectType()),
			boshtbl.NewValueString(e.ObjectName()),
			boshtbl.NewValueString(e.Task()),
			boshtbl.NewValueString(e.Deployment()),
			boshtbl.NewValueString(e.Instance()),

			boshtbl.NewValueString(fmt.Sprintf("%v", e.Context())),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
