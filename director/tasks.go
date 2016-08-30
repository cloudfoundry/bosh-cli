package director

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TaskImpl struct {
	client Client

	id        int
	createdAt time.Time

	state          string
	user           string
	deploymentName string

	description string
	result      string
}

func (t TaskImpl) ID() int              { return t.id }
func (t TaskImpl) CreatedAt() time.Time { return t.createdAt }

func (t TaskImpl) State() string { return t.state }

func (t TaskImpl) IsError() bool {
	return t.state == "error" || t.state == "timeout" || t.state == "cancelled"
}

func (t TaskImpl) User() string           { return t.user }
func (t TaskImpl) DeploymentName() string { return t.deploymentName }

func (t TaskImpl) Description() string { return t.description }
func (t TaskImpl) Result() string      { return t.result }

func (t TaskImpl) Cancel() error { return t.client.CancelTask(t.id) }

type TaskResp struct {
	ID        int   // 165
	Timestamp int64 // 1440318199

	State      string // e.g. "queued", "processing", "done", "error", "cancelled"
	User       string // e.g. "admin"
	Deployment string

	Description string // e.g. "create release"
	Result      string // e.g. "Created release `bosh-ui/0+dev.17'"
}

func NewTaskFromResp(client Client, r TaskResp) TaskImpl {
	return TaskImpl{
		client: client,

		id:        r.ID,
		createdAt: time.Unix(r.Timestamp, 0).UTC(),

		state:          r.State,
		user:           r.User,
		deploymentName: r.Deployment,

		description: r.Description,
		result:      r.Result,
	}
}

func (d DirectorImpl) CurrentTasks(includeAll bool) ([]Task, error) {
	tasks := []Task{}

	taskResps, err := d.client.CurrentTasks(includeAll)
	if err != nil {
		return tasks, err
	}

	for _, r := range taskResps {
		tasks = append(tasks, NewTaskFromResp(d.client, r))
	}

	return tasks, nil
}

func (d DirectorImpl) RecentTasks(limit int, includeAll bool) ([]Task, error) {
	tasks := []Task{}

	taskResps, err := d.client.RecentTasks(limit, includeAll)
	if err != nil {
		return tasks, err
	}

	for _, r := range taskResps {
		tasks = append(tasks, NewTaskFromResp(d.client, r))
	}

	return tasks, nil
}

func (d DirectorImpl) FindTask(id int) (Task, error) {
	taskResp, err := d.client.Task(id)
	if err != nil {
		return TaskImpl{}, err
	}

	return NewTaskFromResp(d.client, taskResp), nil
}

func (t TaskImpl) EventOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "event", taskReporter)
}

func (t TaskImpl) CPIOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "cpi", taskReporter)
}

func (t TaskImpl) DebugOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "debug", taskReporter)
}

func (t TaskImpl) ResultOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "result", taskReporter)
}

func (t TaskImpl) RawOutput(taskReporter TaskReporter) error {
	return t.client.TaskOutput(t.id, "raw", taskReporter)
}

func (c Client) CurrentTasks(includeAll bool) ([]TaskResp, error) {
	var tasks []TaskResp

	path := "/tasks?state=processing,cancelling,queued&verbose=" +
		c.taskVerbosity(includeAll)

	err := c.clientRequest.Get(path, &tasks)
	if err != nil {
		return tasks, bosherr.WrapErrorf(err, "Finding current tasks")
	}

	return tasks, nil
}

func (c Client) RecentTasks(limit int, includeAll bool) ([]TaskResp, error) {
	var tasks []TaskResp

	path := fmt.Sprintf("/tasks?limit=%d&verbose=%s",
		limit, c.taskVerbosity(includeAll))

	err := c.clientRequest.Get(path, &tasks)
	if err != nil {
		return tasks, bosherr.WrapErrorf(err, "Finding recent tasks")
	}

	return tasks, nil
}

func (c Client) taskVerbosity(includeAll bool) string {
	if includeAll {
		return "2"
	}
	return "1"
}

func (c Client) Task(id int) (TaskResp, error) {
	var task TaskResp

	err := c.clientRequest.Get(fmt.Sprintf("/tasks/%d", id), &task)
	if err != nil {
		return task, bosherr.WrapErrorf(err, "Finding task '%d'", id)
	}

	return task, nil
}

func (c Client) TaskOutput(id int, type_ string, taskReporter TaskReporter) error {
	err := c.taskClientRequest.WaitForCompletion(id, type_, taskReporter)
	if err != nil {
		return bosherr.WrapErrorf(err, "Capturing task '%d' output", id)
	}

	return nil
}

func (c Client) CancelTask(id int) error {
	path := fmt.Sprintf("/task/%d", id)

	_, _, err := c.clientRequest.RawDelete(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Cancelling task '%d'", id)
	}

	return nil
}
