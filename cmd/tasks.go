package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type TasksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewTasksCmd(ui boshui.UI, director boshdir.Director) TasksCmd {
	return TasksCmd{ui: ui, director: director}
}

func (c TasksCmd) Run(opts TasksOpts) error {
	filter := boshdir.TasksFilter{
		All:        opts.All,
		Deployment: opts.Deployment,
	}

	if opts.Recent != nil {
		tasks, err := c.director.RecentTasks(*opts.Recent, filter)
		if err != nil {
			return err
		}
		return c.printTable(tasks)
	}

	filter.All = true
	tasks, err := c.director.CurrentTasks(filter)
	if err != nil {
		return err
	}
	return c.printTable(tasks)
}

func (c TasksCmd) printTable(tasks []boshdir.Task) error {
	table := boshtbl.Table{
		Content: "tasks",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("State"),
			boshtbl.NewHeader("Started At"),
			boshtbl.NewHeader("Finished At"),
			boshtbl.NewHeader("User"),
			boshtbl.NewHeader("Deployment"),
			boshtbl.NewHeader("Description"),
			boshtbl.NewHeader("Result"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0}},
	}

	for _, t := range tasks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueInt(t.ID()),
			boshtbl.ValueFmt{
				V:     boshtbl.NewValueString(t.State()),
				Error: t.IsError(),
			},
			boshtbl.NewValueTime(t.StartedAt()),
			sanitizeFinishedAtDisplayValue(t),
			boshtbl.NewValueString(t.User()),
			boshtbl.NewValueString(t.DeploymentName()),
			boshtbl.NewValueString(t.Description()),
			boshtbl.NewValueString(t.Result()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}

func sanitizeFinishedAtDisplayValue(task boshdir.Task) boshtbl.Value {
	if task.FinishedAt().UnixNano() == 0 {
		return boshtbl.NewValueString("-")
	}
	return boshtbl.NewValueTime(task.FinishedAt())
}
