package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type DeploymentRepo interface {
	UpdateCurrent(manifestSHA1 string) error
	FindCurrent() (manifestSHA1 string, found bool, err error)
}

type deploymentRepo struct {
	configService DeploymentConfigService
}

func NewDeploymentRepo(configService DeploymentConfigService) deploymentRepo {
	return deploymentRepo{
		configService: configService,
	}
}

func (r deploymentRepo) FindCurrent() (string, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return "", false, bosherr.WrapError(err, "Loading existing config")
	}

	currentManifestSHA1 := config.CurrentManifestSHA1
	if currentManifestSHA1 != "" {
		return currentManifestSHA1, true, nil
	}

	return "", false, nil
}

func (r deploymentRepo) UpdateCurrent(manifestSHA1 string) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	config.CurrentManifestSHA1 = manifestSHA1

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}
