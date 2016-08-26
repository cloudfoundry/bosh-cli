package director

func (d DeploymentImpl) InstanceInfos() ([]VMInfo, error) {
	infos, err := d.client.DeploymentInstanceInfos(d.name)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func (c Client) DeploymentInstanceInfos(deploymentName string) ([]VMInfo, error) {
	return c.deploymentResourceInfos(deploymentName, "instances")
}
