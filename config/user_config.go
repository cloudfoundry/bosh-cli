package config

import "path"

type UserConfig struct {
	DeploymentManifestPath string `json:"deployment"`
}

func (c UserConfig) DeploymentConfigPath() string {
	return path.Join(path.Dir(c.DeploymentManifestPath), "deployment.json")
}
