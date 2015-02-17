package config

import (
	"path"
)

type UserConfig struct {
	DeploymentManifestPath string `json:"deployment"`
}

func (c UserConfig) IsDeploymentSet() bool {
	return c.DeploymentManifestPath != ""
}

func (c UserConfig) DeploymentConfigPath() string {
	return path.Join(path.Dir(c.DeploymentManifestPath), "deployment.json")
}

func (c UserConfig) LegacyDeploymentConfigPath() string {
	return path.Join(path.Dir(c.DeploymentManifestPath), "bosh-deployments.yml")
}
