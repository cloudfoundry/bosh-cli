package director

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Errand struct {
	Name string // e.g. "acceptance-tests"
}

type ErrandResult struct {
	InstanceGroup string
	InstanceID    string

	ExitCode int

	Stdout string
	Stderr string

	LogsBlobstoreID string
	LogsSHA1        string
}

type ErrandRunResp struct {
	Instance struct {
		Group string `json:"group"`
		ID    string `json:"id"`
	} `json:"instance"`

	ExitCode int `json:"exit_code"`

	Stdout string
	Stderr string

	Logs struct {
		BlobstoreID string `json:"blobstore_id"`
		SHA1        string `json:"sha1"`
	} `json:"logs"`
}

func (d DeploymentImpl) Errands() ([]Errand, error) {
	return d.client.Errands(d.name)
}

func (d DeploymentImpl) RunErrand(name string, keepAlive bool, whenChanged bool, slugs []InstanceGroupOrInstanceSlug) ([]ErrandResult, error) {
	resp, err := d.client.RunErrand(d.name, name, keepAlive, whenChanged, slugs)
	if err != nil {
		return []ErrandResult{}, err
	}

	return parseErrandResults(resp), nil
}

func (d DeploymentImpl) StartErrand(name string, keepAlive bool, whenChanged bool, slugs []InstanceGroupOrInstanceSlug) (int, error) {
	return d.client.StartErrand(d.name, name, keepAlive, whenChanged, slugs)
}

func (d DeploymentImpl) FetchTaskOutputChunk(taskID, offset int, type_ string) ([]byte, int, error) {
	return d.client.FetchTaskOutputChunk(taskID, offset, type_)
}

func (d DeploymentImpl) TaskState(taskID int) (string, error) {
	return d.client.TaskState(taskID)
}

func (d DeploymentImpl) WaitForErrandResult(taskID int) ([]ErrandResult, error) {
	resp, err := d.client.WaitForErrandResult(taskID)
	if err != nil {
		return []ErrandResult{}, err
	}

	return parseErrandResults(resp), nil
}

func parseErrandResults(resp []ErrandRunResp) []ErrandResult {
	var result []ErrandResult

	for _, value := range resp {
		errandResult := ErrandResult{
			InstanceGroup: value.Instance.Group,
			InstanceID:    value.Instance.ID,

			ExitCode: value.ExitCode,

			Stdout: value.Stdout,
			Stderr: value.Stderr,

			LogsBlobstoreID: value.Logs.BlobstoreID,
			LogsSHA1:        value.Logs.SHA1,
		}
		result = append(result, errandResult)
	}

	return result
}

func (c Client) Errands(deploymentName string) ([]Errand, error) {
	var errands []Errand

	if len(deploymentName) == 0 {
		return errands, bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/errands", deploymentName)

	err := c.clientRequest.Get(path, &errands)
	if err != nil {
		return errands, bosherr.WrapErrorf(err, "Finding errands")
	}

	return errands, nil
}

func (c Client) errandRequestBody(deploymentName, name string, keepAlive bool, whenChanged bool, instanceSlugs []InstanceGroupOrInstanceSlug) (string, []byte, error) {
	if len(deploymentName) == 0 {
		return "", nil, bosherr.Error("Expected non-empty deployment name")
	}

	if len(name) == 0 {
		return "", nil, bosherr.Error("Expected non-empty errand name")
	}

	path := fmt.Sprintf("/deployments/%s/errands/%s/runs", deploymentName, name)

	instances := []InstanceFilter{}
	for _, slug := range instanceSlugs {
		instances = append(instances, slug.DirectorHash())
	}

	body := map[string]interface{}{
		"keep-alive":   keepAlive,
		"when-changed": whenChanged,
		"instances":    instances,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", nil, bosherr.WrapErrorf(err, "Marshaling request body")
	}

	return path, reqBody, nil
}

func (c Client) RunErrand(deploymentName, name string, keepAlive bool, whenChanged bool, instanceSlugs []InstanceGroupOrInstanceSlug) ([]ErrandRunResp, error) {
	path, reqBody, err := c.errandRequestBody(deploymentName, name, keepAlive, whenChanged, instanceSlugs)
	if err != nil {
		return nil, err
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	resultBytes, err := c.taskClientRequest.PostResult(path, reqBody, setHeaders)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Running errand '%s'", name)
	}

	return parseErrandRunResps(resultBytes)
}

func (c Client) StartErrand(deploymentName, name string, keepAlive bool, whenChanged bool, instanceSlugs []InstanceGroupOrInstanceSlug) (int, error) {
	path, reqBody, err := c.errandRequestBody(deploymentName, name, keepAlive, whenChanged, instanceSlugs)
	if err != nil {
		return 0, err
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	taskID, err := c.taskClientRequest.PostTaskID(path, reqBody, setHeaders)
	if err != nil {
		return 0, bosherr.WrapErrorf(err, "Starting errand '%s'", name)
	}

	return taskID, nil
}

func (c Client) WaitForErrandResult(taskID int) ([]ErrandRunResp, error) {
	resultBytes, err := c.taskClientRequest.WaitForTaskResultOnly(taskID)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Waiting for errand task '%d'", taskID)
	}

	return parseErrandRunResps(resultBytes)
}

func parseErrandRunResps(resultBytes []byte) ([]ErrandRunResp, error) {
	var resp []ErrandRunResp

	dec := json.NewDecoder(strings.NewReader(string(resultBytes)))

	for {
		var errandRunResponse ErrandRunResp
		if err := dec.Decode(&errandRunResponse); err == io.EOF {
			break
		} else if err != nil {
			return nil, bosherr.WrapErrorf(err, "Unmarshaling errand result")
		}
		resp = append(resp, errandRunResponse)
	}

	return resp, nil
}
