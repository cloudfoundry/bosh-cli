package config

import "path"

type Config struct {
	Deployment     string `json:"deployment"`
	ContainingDir  string `json:"-"`
	DeploymentUUID string `json:"-"`
}

func (c Config) DeploymentFile() string {
	return path.Join(path.Dir(c.Deployment), "deployment.json")
}

func (c Config) BlobstorePath() string {
	return path.Join(c.ContainingDir, ".bosh_micro", c.DeploymentUUID, "blobs")
}

func (c Config) IndexPath() string {
	return path.Join(c.ContainingDir, ".bosh_micro", c.DeploymentUUID, "index.json")
}

func (c Config) PackagesPath() string {
	return path.Join(c.ContainingDir, ".bosh_micro", c.DeploymentUUID, "packages")
}
