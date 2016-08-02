package director

import (
	//"fmt"
	"time"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"fmt"
)

type EventImpl struct {
	client Client

	id        int
	timestamp time.Time
	user  string
	action string
	objectType string
	objectName string
	task string
	deployment string
	instance string
	context map[string]interface{}
}

type EventResp struct {
	Id        int   // 165
	Timestamp int64 // 1440318199
	User  string // e.g. "admin"
	Action string
	ObjectType string
	ObjectName string
	Task string
	Deployment string
	Instance string
	Context map[string]interface{}
}

func (e EventImpl) Id() int          			 { return e.id }
func (e EventImpl) Timestamp() time.Time   		 { return e.timestamp }
func (e EventImpl) User() string         		 { return e.user }
func (e EventImpl) Action() string       		 { return e.action }
func (e EventImpl) ObjectType() string  		 { return e.objectType }
func (e EventImpl) ObjectName() string  		 { return e.objectName }
func (e EventImpl) Task() string       			 { return e.task }
func (e EventImpl) Deployment() string  		 { return e.deployment }
func (e EventImpl) Instance() string    		 { return e.instance }
func (e EventImpl) Context() map[string]interface{}      { return e.context }

func NewEventFromResp(client Client, r EventResp) EventImpl {
	return EventImpl{
		client: client,

		id:        r.Id,
		timestamp: time.Unix(r.Timestamp, 0).UTC(),
		user:  r.User,
		action: r.Action,
		objectType: r.ObjectType,
		objectName: r.ObjectName,
		task: r.Task,
		deployment: r.Deployment,
		instance: r.Instance,
		context: r.Context,
	}
}


func (d DirectorImpl) Events(beforeId int, before time.Time, after time.Time, deployment string, task string, instance string) ([]Event, error) {
	events := []Event{}

	eventResps, err := d.client.Events(beforeId, before, after, deployment, task, instance)
	if err != nil {
		return events, err
	}

	for _, r := range eventResps {
		events = append(events, NewEventFromResp(d.client, r))
	}

	return events, nil
}

func (c Client) Events(beforeId int, before time.Time, after time.Time, deployment string, task string, instance string) ([]EventResp, error) {
	var events []EventResp

	path := fmt.Sprintf("/events/?deployment=%s", "deployment")
	path = ""
	println(path)

	err := c.clientRequest.Get(path, &events)
	if err != nil {
		return events, bosherr.WrapErrorf(err, "Finding events")
	}

	return events, nil
}
