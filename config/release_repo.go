package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
	"github.com/cloudfoundry/bosh-micro-cli/release"
)

// ReleaseRepo persists releases metadata
type ReleaseRepo interface {
	List() ([]ReleaseRecord, error)
	Update([]release.Release) error
}

type releaseRepo struct {
	configService DeploymentConfigService
	uuidGenerator boshuuid.Generator
}

func NewReleaseRepo(configService DeploymentConfigService, uuidGenerator boshuuid.Generator) ReleaseRepo {
	return releaseRepo{
		configService: configService,
		uuidGenerator: uuidGenerator,
	}
}

func (r releaseRepo) Update(releases []release.Release) error {
	var newRecordIDs []string
	var newRecords []ReleaseRecord

	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	for _, release := range releases {
		newRecord := ReleaseRecord{
			Name:    release.Name(),
			Version: release.Version(),
		}
		newRecord.ID, err = r.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating release id")
		}
		newRecords = append(newRecords, newRecord)
		newRecordIDs = append(newRecordIDs, newRecord.ID)
	}

	config.CurrentReleaseIDs = newRecordIDs
	config.Releases = newRecords
	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Updating current release record")
	}
	return nil
}

func (r releaseRepo) List() ([]ReleaseRecord, error) {
	config, err := r.configService.Load()
	if err != nil {
		return []ReleaseRecord{}, bosherr.WrapError(err, "Loading existing config")
	}
	return config.Releases, nil
}
