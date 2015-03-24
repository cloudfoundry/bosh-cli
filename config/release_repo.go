package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

// ReleaseRepo persists releases metadata
type ReleaseRepo interface {
	UpdateCurrent(recordIDs []string) error
	FindCurrent() ([]ReleaseRecord, bool, error)
	Save(name, version string) (ReleaseRecord, error)
	Find(name, version string) (ReleaseRecord, bool, error)
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

func (r releaseRepo) Save(name, version string) (ReleaseRecord, error) {
	config, err := r.configService.Load()
	if err != nil {
		return ReleaseRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Releases
	if records == nil {
		records = []ReleaseRecord{}
	}

	newRecord := ReleaseRecord{
		Name:    name,
		Version: version,
	}
	newRecord.ID, err = r.uuidGenerator.Generate()
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Generating release id")
	}

	for _, oldRecord := range records {
		if oldRecord.Name == newRecord.Name && oldRecord.Version == newRecord.Version {
			return oldRecord, bosherr.Errorf("Failed to save release record '%s' (duplicate name/version), existing record found '%s'", newRecord, oldRecord)
		}
	}

	records = append(records, newRecord)
	config.Releases = records

	err = r.configService.Save(config)
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Saving new config")
	}
	return newRecord, nil
}

func (r releaseRepo) Find(name, version string) (ReleaseRecord, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return ReleaseRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Releases
	if records == nil {
		return ReleaseRecord{}, false, nil
	}

	for _, oldRecord := range records {
		if oldRecord.Name == name && oldRecord.Version == version {
			return oldRecord, true, nil
		}
	}
	return ReleaseRecord{}, false, nil
}

func (r releaseRepo) FindCurrent() ([]ReleaseRecord, bool, error) {
	config, err := r.configService.Load()
	currentReleaseRecords := []ReleaseRecord{}
	if err != nil {
		return currentReleaseRecords, false, bosherr.WrapError(err, "Loading existing config")
	}

	currentReleaseIDs := config.CurrentReleaseIDs

	for _, currentReleaseID := range currentReleaseIDs {
		for _, oldRecord := range config.Releases {
			if oldRecord.ID == currentReleaseID {
				currentReleaseRecords = append(currentReleaseRecords, oldRecord)
			}
		}
	}

	found := len(currentReleaseIDs) > 0 && len(currentReleaseRecords) == len(currentReleaseIDs)

	return currentReleaseRecords, found, nil
}

func (r releaseRepo) UpdateCurrent(recordIDs []string) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	for _, recordID := range recordIDs {
		found := false
		for _, oldRecord := range config.Releases {
			if oldRecord.ID == recordID {
				found = true
				break
			}
		}
		if !found {
			return bosherr.Errorf("Verifying release record exists with id '%s'", recordID)
		}
	}

	config.CurrentReleaseIDs = recordIDs

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}
