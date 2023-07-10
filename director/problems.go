package director

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Problem struct {
	ID int // e.g. 4

	Type          string // e.g. "unresponsive_agent"
	Description   string // e.g. "api/1 (5efd2cb8-d73b-4e45-6df4-58f5dd5ec2ec) is not responding"
	InstanceGroup string `json:"instance_group"` // e.g. "database"

	Data        interface{}
	Resolutions []ProblemResolution
}

var skipResolutionName string = "ignore"

var ProblemResolutionDefault ProblemResolution = ProblemResolution{}
var ProblemResolutionSkip ProblemResolution = ProblemResolution{
	Name: &skipResolutionName,
	Plan: "Skip for now",
}

type ProblemResolution struct {
	Name *string `json:"name"` // e.g. "Skip for now", "Recreate VM"
	Plan string  `json:"plan"` // e.g. "ignore", "reboot_vm"
}

type ProblemAnswer struct {
	ProblemID  int               `yaml:"problem_id"`
	Resolution ProblemResolution `yaml:"resolution"`
}

func (d DeploymentImpl) ScanForProblems() ([]Problem, error) {
	err := d.client.ScanForProblems(d.name)
	if err != nil {
		return nil, err
	}

	return d.client.ListProblems(d.name)
}

func (d DeploymentImpl) ResolveProblems(answers []ProblemAnswer, overrides map[string]string) error {
	return d.client.ResolveProblems(d.name, answers, overrides)
}

func (c Client) ScanForProblems(deploymentName string) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/scans", deploymentName)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, err := c.taskClientRequest.PostResult(path, nil, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Performing a scan on deployment '%s'", deploymentName)
	}

	return nil
}

func (c Client) ListProblems(deploymentName string) ([]Problem, error) {
	var probs []Problem

	if len(deploymentName) == 0 {
		return probs, bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/problems", deploymentName)

	err := c.clientRequest.Get(path, &probs)
	if err != nil {
		return probs, bosherr.WrapErrorf(
			err, "Listing problems for deployment '%s'", deploymentName)
	}

	return probs, nil
}

type problemResolutionBody struct {
	Resolutions          map[string]*string `json:"resolutions"`
	MaxInFlightOverrides map[string]string  `json:"max_in_flight_overrides,omitempty"`
}

func (c Client) ResolveProblems(deploymentName string, answers []ProblemAnswer, overrides map[string]string) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s/problems", deploymentName)

	body := &problemResolutionBody{
		Resolutions:          make(map[string]*string),
		MaxInFlightOverrides: overrides,
	}

	for _, ans := range answers {
		body.Resolutions[strconv.Itoa(ans.ProblemID)] = ans.Resolution.Name
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, err = c.taskClientRequest.PutResult(path, reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Resolving problems for deployment '%s'", deploymentName)
	}

	return nil
}
