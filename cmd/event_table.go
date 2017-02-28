package cmd

import (
	"fmt"
	"sort"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type EventTable struct {
	Event boshdir.Event
	UI    boshui.UI
}

func (t EventTable) Print() {
	id := t.Event.ID()
	if t.Event.ParentID() != "" {
		id += " <- " + t.Event.ParentID()
	}
	table := boshtbl.Table{
		Rows: [][]boshtbl.Value{
			{
				boshtbl.NewValueString("ID"),
				boshtbl.NewValueString(id),
			},
			{
				boshtbl.NewValueString("Time"),
				boshtbl.NewValueTime(t.Event.Timestamp()),
			},
		},
	}

	if len(t.Event.User()) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("User"),
			boshtbl.NewValueString(t.Event.User()),
		})
	}

	table.Rows = append(table.Rows, []boshtbl.Value{
		boshtbl.NewValueString("Action"),
		boshtbl.NewValueString(t.Event.Action()),
	})

	table.Rows = append(table.Rows, []boshtbl.Value{
		boshtbl.NewValueString("Object Type"),
		boshtbl.NewValueString(t.Event.ObjectType()),
	})

	if len(t.Event.ObjectName()) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Object ID"),
			boshtbl.NewValueString(t.Event.ObjectName()),
		})
	}

	if len(t.Event.TaskID()) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Task ID"),
			boshtbl.NewValueString(t.Event.TaskID()),
		})
	}

	if len(t.Event.DeploymentName()) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Deployment"),
			boshtbl.NewValueString(t.Event.DeploymentName()),
		})
	}

	if len(t.Event.Instance()) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Instance"),
			boshtbl.NewValueString(t.Event.Instance()),
		})
	}

	if len(t.Event.Context()) > 0 {
		desc := []string{}

		for name, value := range t.Event.Context() {
			desc = append(desc, fmt.Sprintf("%s: %s", name, value))
		}

		sort.Sort(EventContextSorting(desc))

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Context"),
			boshtbl.NewValueStrings(desc),
		})
	}

	if len(t.Event.Error()) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Error"),
			boshtbl.NewValueString(t.Event.Error()),
		})
	}

	t.UI.PrintTable(table)
}

type EventContextSorting []string

func (s EventContextSorting) Len() int           { return len(s) }
func (s EventContextSorting) Less(i, j int) bool { return s[i] < s[j] }
func (s EventContextSorting) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
