package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type TasksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewTasksCmd(ui boshui.UI, director boshdir.Director) TasksCmd {
	return TasksCmd{ui: ui, director: director}
}

func (c TasksCmd) Run(opts TasksOpts) error {
	if opts.Recent != nil {
		return c.printTable(c.director.RecentTasks(*opts.Recent, opts.All))
	}
	return c.printTable(c.director.CurrentTasks(opts.All))
}

func (c TasksCmd) printTable(tasks []boshdir.Task, err error) error {
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "tasks",
		Header:  []string{"#", "State", "Created At", "User", "Deployment", "Description", "Result"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0}},
	}

	for _, t := range tasks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueInt(t.ID()),
			boshtbl.ValueFmt{
				V:     boshtbl.NewValueString(t.State()),
				Error: t.IsError(),
			},
			boshtbl.NewValueTime(t.CreatedAt()),
			boshtbl.NewValueString(t.User()),
			boshtbl.NewValueString(t.DeploymentName()),
			boshtbl.NewValueString(t.Description()),
			boshtbl.NewValueString(t.Result()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
