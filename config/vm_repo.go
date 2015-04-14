package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type VMRepo interface {
	FindCurrent() (cid string, found bool, err error)
	UpdateCurrent(cid string) error
	ClearCurrent() error
}

type vMRepo struct {
	configService DeploymentStateService
}

func NewVMRepo(configService DeploymentStateService) VMRepo {
	return vMRepo{
		configService: configService,
	}
}

func (r vMRepo) FindCurrent() (string, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return "", false, bosherr.WrapError(err, "Loading existing config")
	}

	currentVMCID := config.CurrentVMCID
	if currentVMCID != "" {
		return currentVMCID, true, nil
	}

	return "", false, nil
}

func (r vMRepo) UpdateCurrent(cid string) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	config.CurrentVMCID = cid

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (r vMRepo) ClearCurrent() error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	config.CurrentVMCID = ""

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}
