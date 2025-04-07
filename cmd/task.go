package cmd

import (
	"errors"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshuit "github.com/cloudfoundry/bosh-cli/v7/ui/task"
)

type TaskCmd struct {
	eventsTaskReporter boshuit.Reporter
	plainTaskReporter  boshuit.Reporter
	director           boshdir.Director
}

func NewTaskCmd(
	eventsTaskReporter boshuit.Reporter,
	plainTaskReporter boshuit.Reporter,
	director boshdir.Director,
) TaskCmd {
	return TaskCmd{
		eventsTaskReporter: eventsTaskReporter,
		plainTaskReporter:  plainTaskReporter,
		director:           director,
	}
}

func (c TaskCmd) Run(opts TaskOpts) error {
	var task boshdir.Task

	var err error

	if opts.Args.ID == 0 {
		filter := boshdir.TasksFilter{
			All:        opts.All,
			Deployment: opts.Deployment,
		}
		tasks, err := c.director.CurrentTasks(filter)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			return errors.New("No task found") //nolint:staticcheck
		}

		task = tasks[0]
	} else {
		task, err = c.director.FindTask(opts.Args.ID)
		if err != nil {
			return err
		}
	}

	switch {
	case opts.Event:
		err = task.EventOutput(c.plainTaskReporter)
	case opts.CPI:
		err = task.CPIOutput(c.plainTaskReporter)
	case opts.Debug:
		err = task.DebugOutput(c.plainTaskReporter)
	case opts.Result:
		err = task.ResultOutput(c.plainTaskReporter)
	default:
		err = task.EventOutput(c.eventsTaskReporter)
	}

	return err
}
