package config

import "path"

type DeploymentConfig struct {
	ContainingDir  string
	DeploymentUUID string
	CurrentVMCID   string
	CurrentDiskID string
	Disks     []DiskRecord
	Stemcells []StemcellRecord
}

func (c DeploymentConfig) BlobstorePath() string {
	return path.Join(c.deploymentDir(), "blobs")
}

func (c DeploymentConfig) CompiledPackagedIndexPath() string {
	return path.Join(c.deploymentDir(), "compiled_packages.json")
}

func (c DeploymentConfig) TemplatesIndexPath() string {
	return path.Join(c.deploymentDir(), "templates.json")
}

func (c DeploymentConfig) PackagesPath() string {
	return path.Join(c.deploymentDir(), "packages")
}

func (c DeploymentConfig) JobsPath() string {
	return path.Join(c.deploymentDir(), "jobs")
}

func (c DeploymentConfig) deploymentDir() string {
	return path.Join(c.ContainingDir, c.DeploymentUUID)
}
