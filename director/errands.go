package director

import (
	"encoding/json"
	"fmt"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Errand struct {
	Name string // e.g. "acceptance-tests"
}

type ErrandResult struct {
	ExitCode int

	Stdout string
	Stderr string

	LogsBlobstoreID string
	LogsSHA1        string
}

type ErrandRunResp struct {
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

func (d DeploymentImpl) RunErrand(name string, keepAlive bool) (ErrandResult, error) {
	resp, err := d.client.RunErrand(d.name, name, keepAlive)
	if err != nil {
		return ErrandResult{}, err
	}

	result := ErrandResult{
		ExitCode: resp.ExitCode,

		Stdout: resp.Stdout,
		Stderr: resp.Stderr,

		LogsBlobstoreID: resp.Logs.BlobstoreID,
		LogsSHA1:        resp.Logs.SHA1,
	}

	return result, nil
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

func (c Client) RunErrand(deploymentName, name string, keepAlive bool) (ErrandRunResp, error) {
	var resp ErrandRunResp

	if len(deploymentName) == 0 {
		return resp, bosherr.Error("Expected non-empty deployment name")
	}

	if len(name) == 0 {
		return resp, bosherr.Error("Expected non-empty errand name")
	}

	path := fmt.Sprintf("/deployments/%s/errands/%s/runs", deploymentName, name)

	body := map[string]bool{"keep-alive": keepAlive}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	resultBytes, err := c.taskClientRequest.PostResult(path, reqBody, setHeaders)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Running errand '%s'", name)
	}

	err = json.Unmarshal(resultBytes, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Unmarshaling errand result")
	}

	return resp, nil
}
