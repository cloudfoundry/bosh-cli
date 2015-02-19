package installation

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type TargetProvider interface {
	NewTarget() (Target, error)
}

type targetProvider struct {
	deploymentConfigService bmconfig.DeploymentConfigService
	uuidGenerator           boshuuid.Generator
	installationsRootPath   string
}

func NewTargetProvider(
	deploymentConfigService bmconfig.DeploymentConfigService,
	uuidGenerator boshuuid.Generator,
	installationsRootPath string,
) TargetProvider {
	return &targetProvider{
		deploymentConfigService: deploymentConfigService,
		uuidGenerator:           uuidGenerator,
		installationsRootPath:   installationsRootPath,
	}
}

func (p *targetProvider) NewTarget() (Target, error) {
	deploymentConfig, err := p.deploymentConfigService.Load()
	if err != nil {
		return Target{}, bosherr.WrapError(err, "Loading deployment config")
	}

	installationID := deploymentConfig.InstallationID
	if installationID == "" {
		installationID, err = p.uuidGenerator.Generate()
		if err != nil {
			return Target{}, bosherr.WrapError(err, "Generating installation ID")
		}

		deploymentConfig.InstallationID = installationID
		err := p.deploymentConfigService.Save(deploymentConfig)
		if err != nil {
			return Target{}, bosherr.WrapError(err, "Saving deployment config")
		}
	}

	return NewTarget(filepath.Join(p.installationsRootPath, installationID)), nil
}
