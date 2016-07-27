package director

import (
	"fmt"
	"net/http"
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentDiffResponse struct {
	Context map[string]interface{} `json:"context"`
	Diff    [][]interface{}        `json:"diff"`
}

type DiffLines [][]interface{}

func (d DeploymentImpl) Diff(manifest []byte, doNotRedact bool) (DiffLines, error) {
	resp, err := d.client.Diff(manifest, d.name, doNotRedact)
	if err != nil {
		return DiffLines{}, err
	}

	return DiffLines(resp.Diff), nil
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

	path := fmt.Sprintf("/deployments/%s/diff?%s", deploymentName, query.Encode())

	var resp DeploymentDiffResponse

	err := c.clientRequest.Post(path, manifest, setHeaders, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Fetching diff result")
	}

	return resp, nil
}
