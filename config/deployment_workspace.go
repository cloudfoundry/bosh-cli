package config

import (
	"path/filepath"
)

type DeploymentWorkspace struct {
	containingDir  string
	deploymentUUID string
}

func NewDeploymentWorkspace(containingDir, deploymentUUID string) DeploymentWorkspace {
	return DeploymentWorkspace{
		containingDir:  containingDir,
		deploymentUUID: deploymentUUID,
	}
}

func (w DeploymentWorkspace) DeploymentUUID() string {
	return w.deploymentUUID
}

func (w DeploymentWorkspace) BlobstorePath() string {
	return filepath.Join(w.deploymentDir(), "blobs")
}

func (w DeploymentWorkspace) CompiledPackagedIndexPath() string {
	return filepath.Join(w.deploymentDir(), "compiled_packages.json")
}

func (w DeploymentWorkspace) TemplatesIndexPath() string {
	return filepath.Join(w.deploymentDir(), "templates.json")
}

func (w DeploymentWorkspace) PackagesPath() string {
	return filepath.Join(w.deploymentDir(), "packages")
}

func (w DeploymentWorkspace) JobsPath() string {
	return filepath.Join(w.deploymentDir(), "jobs")
}

func (w DeploymentWorkspace) deploymentDir() string {
	return filepath.Join(w.containingDir, w.deploymentUUID)
}
