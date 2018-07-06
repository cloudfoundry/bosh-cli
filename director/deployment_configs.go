package director

import (
	"encoding/json"
	"fmt"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentConfigConfig struct {
	Id   string
	Type string
	Name string
}

type DeploymentConfig struct {
	Config DeploymentConfigConfig
}

func (d DirectorImpl) ListDeploymentConfigs(name string) ([]DeploymentConfig, bool, error) {
	return d.client.ListDeploymentConfigs(name)
}

func (c Client) ListDeploymentConfigs(name string) ([]DeploymentConfig, bool, error) {
	var deps []DeploymentConfig

	path := fmt.Sprintf("/deployment_configs?deployment[]=%s", name)
	respBody, response, err := c.clientRequest.RawGet(path, nil, nil)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			// endpoint couldn't be found => return empty array for compatibility with old directors
			return deps, false, nil
		}
		return deps, false, err
	}

	err = json.Unmarshal(respBody, &deps)
	if err != nil {
		return deps, false, bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return deps, true, nil
}
