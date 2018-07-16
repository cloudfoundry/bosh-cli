package director

import (
	"encoding/json"
	"fmt"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentConfigProperties struct {
	Id   int
	Type string
	Name string
}

type DeploymentConfig struct {
	Config DeploymentConfigProperties
}

type DeploymentConfigs struct {
	Configs []DeploymentConfig
}

func (d DeploymentConfigs) GetConfig(idx int) DeploymentConfigProperties {
	return d.Configs[idx].Config
}

func (d DeploymentConfigs) GetConfigs() []DeploymentConfigProperties {
	configProperties := make([]DeploymentConfigProperties, len(d.Configs))
	for i, cp := range d.Configs {
		configProperties[i] = cp.Config
	}
	return configProperties
}

func (d DirectorImpl) ListDeploymentConfigs(name string) (DeploymentConfigs, error) {
	return d.client.listDeploymentConfigs(name)
}

func (c Client) listDeploymentConfigs(name string) (DeploymentConfigs, error) {
	var deps DeploymentConfigs

	path := fmt.Sprintf("/deployment_configs?deployment[]=%s", name)
	respBody, response, err := c.clientRequest.RawGet(path, nil, nil)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			// endpoint couldn't be found => return empty array for compatibility with old directors
			return deps, nil
		}
		return deps, err
	}

	err = json.Unmarshal(respBody, &deps.Configs)
	if err != nil {
		return deps, bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return deps, nil
}
