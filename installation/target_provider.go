package installation

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
)

type TargetProvider interface {
	NewTarget() (Target, error)
}

type targetProvider struct {
	deploymentStateService biconfig.DeploymentStateService
	uuidGenerator          boshuuid.Generator
	installationsRootPath  string
	packageDir             string
}

func NewTargetProvider(
	deploymentStateService biconfig.DeploymentStateService,
	uuidGenerator boshuuid.Generator,
	installationsRootPath string,
	packageDir string,
) TargetProvider {
	return &targetProvider{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
		installationsRootPath:  installationsRootPath,
		packageDir:             packageDir,
	}
}

func (p *targetProvider) NewTarget() (Target, error) {
	deploymentState, err := p.deploymentStateService.Load()
	if err != nil {
		return Target{}, bosherr.WrapError(err, "Loading deployment state")
	}

	installationID := deploymentState.InstallationID
	if installationID == "" {
		installationID, err = p.uuidGenerator.Generate()
		if err != nil {
			return Target{}, bosherr.WrapError(err, "Generating installation ID")
		}

		deploymentState.InstallationID = installationID
		err := p.deploymentStateService.Save(deploymentState)
		if err != nil {
			return Target{}, bosherr.WrapError(err, "Saving deployment state")
		}
	}

	return NewTarget(filepath.Join(p.installationsRootPath, installationID), p.packageDir), nil
}
