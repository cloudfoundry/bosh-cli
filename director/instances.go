package director

import (
	"strings"
)

func (d DeploymentImpl) InstanceInfos() ([]VMInfo, error) {
	infos, err := d.client.DeploymentInstanceInfos(d.name)

	if err != nil && strings.Contains(err.Error(), "404") {
		infos, err = d.client.DeploymentVMInfos(d.name)
	}

	if err != nil {
		return nil, err
	}

	addTimestampToInfos(infos)

	return infos, nil
}

func (c Client) DeploymentInstanceInfos(deploymentName string) ([]VMInfo, error) {
	return c.DeploymentResourceInfos(deploymentName, "instances")
}
