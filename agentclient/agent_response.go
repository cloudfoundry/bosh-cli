package agentclient

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type Response interface {
	GetException() exceptionResponse
}

type SimpleResponse struct {
	Value     string
	Exception exceptionResponse
}

func (r *SimpleResponse) GetException() exceptionResponse {
	return r.Exception
}

type TaskResponse struct {
	Value     interface{}
	Exception exceptionResponse
}

type exceptionResponse struct {
	Message string
}

func (r *TaskResponse) GetException() exceptionResponse {
	return r.Exception
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
		taskState, ok := complexResponse["state"]
		if ok {
			return taskState.(string), nil
		}

		return "", bosherr.New("Failed to parse task state from agent response %#v", r.Value)
	}

	return r.Value.(string), nil
}
