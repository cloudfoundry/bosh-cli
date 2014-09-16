package config

import "path"

type Config struct {
	Deployment     string           `json:"deployment"`
	ContainingDir  string           `json:"-"`
	DeploymentUUID string           `json:"-"`
	Stemcells      []StemcellRecord `json:"-"`
}

func (c Config) DeploymentFile() string {
	return path.Join(path.Dir(c.Deployment), "deployment.json")
}

func (c Config) BlobstorePath() string {
	return path.Join(c.deploymentDir(), "blobs")
}

func (c Config) CompiledPackagedIndexPath() string {
	return path.Join(c.deploymentDir(), "compiled_packages.json")
}

func (c Config) TemplatesIndexPath() string {
	return path.Join(c.deploymentDir(), "templates.json")
}

func (c Config) PackagesPath() string {
	return path.Join(c.deploymentDir(), "packages")
}

func (c Config) JobsPath() string {
	return path.Join(c.deploymentDir(), "jobs")
}

func (c Config) deploymentDir() string {
	return path.Join(c.ContainingDir, ".bosh_micro", c.DeploymentUUID)
}
