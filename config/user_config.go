package config

import "path"

type UserConfig struct {
	DeploymentFile string `json:"deployment"`
}

func (c UserConfig) DeploymentConfigFilePath() string {
	return path.Join(path.Dir(c.DeploymentFile), "deployment.json")
}
