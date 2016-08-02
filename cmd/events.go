package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
	//"time"
)

//TODO: change the member functions.

//type Event struct {
//	Id         int
//	Timestamp  time.Time
//	User       string
//	Action     string
//	ObjectType string
//	ObjectName string
//	Task       string
//	Deployment string
//	Instance   string
//	Context    map[string]interface{}
//}

type EventsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewEventsCmd(ui boshui.UI, director boshdir.Director) EventsCmd {
	return EventsCmd{ui: ui, director: director}
}

func (c EventsCmd) Run(opts EventsOpts) error {
	return c.printTable(c.director.Events(opts.BeforeId, opts.Before, opts.After, opts.Deployment, opts.Task, opts.Instance))
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
			boshtbl.NewValueString("e.Context()"), //TODO: Print context hash
		})
	}

	c.ui.PrintTable(table)

	return nil
}
