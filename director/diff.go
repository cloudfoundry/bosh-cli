package director

import (
	"fmt"
	"net/http"
	gourl "net/url"

	//bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentDiffResponse struct {
	Context map[string]interface{} `json:"context"`
	Diff    [][]interface{}        `json:"diff"`
}

type DeploymentDiff [][]interface{}

func (d DeploymentImpl) Diff(manifest []byte, doNotRedact bool) (DeploymentDiff, error) {
	resp, err := d.client.Diff(manifest, d.name, doNotRedact)
	if err != nil {
		//return DeploymentDiff{}, err
	}

	return DeploymentDiff(resp.Diff), nil
}

func (c Client) Diff(manifest []byte, deploymentName string, doNotRedact bool) (DeploymentDiffResponse, error) {
	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	query := gourl.Values{}

	if doNotRedact {
		query.Add("redact", "false")
	} else {
		query.Add("redact", "true")
	}

	path := fmt.Sprintf("/deployments/%s/diff?%s", deploymentName,query.Encode())

	var response DeploymentDiffResponse

	err := c.clientRequest.Post(path, manifest, setHeaders, &response)
	if err != nil {
		//return response, bosherr.WrapErrorf(err, "making request")
	}

	return response, nil
}
