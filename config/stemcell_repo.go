package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

// StemcellRepo persists stemcells metadata
type StemcellRepo interface {
	Save(stemcell StemcellRecord) error
	Find(name, version string) (StemcellRecord, bool, error)
}

type repo struct {
	configService DeploymentConfigService
}

func NewStemcellRepo(configService DeploymentConfigService) repo {
	return repo{
		configService: configService,
	}
}

func (r repo) Save(newRecord StemcellRecord) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		records = []StemcellRecord{}
	}

	oldRecord, found := r.find(records, newRecord.Name, newRecord.Version)
	if found {
		return bosherr.New("Failed to save stemcell record `%s', existing record found `%s'", newRecord, oldRecord)
	}

	records = append(records, newRecord)
	config.Stemcells = records

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (r repo) Find(name, version string) (StemcellRecord, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return StemcellRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		return StemcellRecord{}, false, nil
	}

	foundRecord, found := r.find(records, name, version)
	return foundRecord, found, nil
}

func (r repo) find(records []StemcellRecord, name, version string) (StemcellRecord, bool) {
	for _, existingRecord := range records {
		if existingRecord.Name == name && existingRecord.Version == version {
			return existingRecord, true
		}
	}
	return StemcellRecord{}, false
}
