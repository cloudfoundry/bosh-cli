package agentclient

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type Response interface {
	Unmarshal([]byte) error
	GetException() exceptionResponse
}

type exceptionResponse struct {
	Message string
}

func (r exceptionResponse) IsEmpty() bool {
	return r == exceptionResponse{}
}

type SimpleTaskResponse struct {
	Value     string
	Exception exceptionResponse
}

func (r *SimpleTaskResponse) GetException() exceptionResponse {
	return r.Exception
}

func (r *SimpleTaskResponse) Unmarshal(message []byte) error {
	return json.Unmarshal(message, r)
}

type StateResponse struct {
	Value     State
	Exception exceptionResponse
}

type State struct {
	JobState string `json:"job_state"`
}

func (r *StateResponse) GetException() exceptionResponse {
	return r.Exception
}

func (r *StateResponse) Unmarshal(message []byte) error {
	return json.Unmarshal(message, r)
}

type TaskResponse struct {
	Value     interface{}
	Exception exceptionResponse
}

func (r *TaskResponse) GetException() exceptionResponse {
	return r.Exception
}

func (r *TaskResponse) Unmarshal(message []byte) error {
	return json.Unmarshal(message, r)
}

func (r *TaskResponse) TaskID() (string, error) {
	complexResponse, ok := r.Value.(map[string]interface{})
	if !ok {
		return "", bosherr.New("Failed to convert agent response to map %#v", r.Value)
	}

	agentTaskID, ok := complexResponse["agent_task_id"]
	if !ok {
		return "", bosherr.New("Failed to parse task id from agent response %#v", r.Value)
	}

	return agentTaskID.(string), nil
}

// TaskState returns the state of the task reported by agent.
//
// Agent response to get_task can be in different format based on task state.
// If task state is running agent responds
// with value as { agent_task_id: "task-id", state: "running" }
// Otherwise the value is a string like "stopped".
func (r *TaskResponse) TaskState() (string, error) {
	complexResponse, ok := r.Value.(map[string]interface{})
	if ok {
		_, ok := complexResponse["agent_task_id"]
		if ok {
			taskState, ok := complexResponse["state"]
			if ok {
				return taskState.(string), nil
			}

			return "", bosherr.New("Failed to parse task state from agent response %#v", r.Value)
		}
	}

	return "finished", nil
}
