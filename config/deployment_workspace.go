package config

import (
	"path/filepath"
)

type DeploymentWorkspace struct {
	containingDir string
	deploymentID  string
}

func NewDeploymentWorkspace(containingDir, deploymentID string) DeploymentWorkspace {
	return DeploymentWorkspace{
		containingDir: containingDir,
		deploymentID:  deploymentID,
	}
}

func (w DeploymentWorkspace) DeploymentID() string {
	return w.deploymentID
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
	return filepath.Join(w.containingDir, w.deploymentID)
}
