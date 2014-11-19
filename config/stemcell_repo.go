package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

// StemcellRepo persists stemcells metadata
type StemcellRepo interface {
	UpdateCurrent(recordID string) error
	FindCurrent() (StemcellRecord, bool, error)
	Save(name, version, cid string) (StemcellRecord, error)
	Find(name, version string) (StemcellRecord, bool, error)
}

type stemcellRepo struct {
	configService DeploymentConfigService
	uuidGenerator boshuuid.Generator
}

func NewStemcellRepo(configService DeploymentConfigService, uuidGenerator boshuuid.Generator) stemcellRepo {
	return stemcellRepo{
		configService: configService,
		uuidGenerator: uuidGenerator,
	}
}

func (r stemcellRepo) Save(name, version, cid string) (StemcellRecord, error) {
	config, err := r.configService.Load()
	if err != nil {
		return StemcellRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		records = []StemcellRecord{}
	}

	newRecord := StemcellRecord{
		Name:    name,
		Version: version,
		CID:     cid,
	}
	newRecord.ID, err = r.uuidGenerator.Generate()
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Generating stemcell id")
	}

	for _, oldRecord := range records {
		if oldRecord.Name == newRecord.Name && oldRecord.Version == newRecord.Version {
			return oldRecord, bosherr.New("Failed to save stemcell record `%s' (duplicate name/version), existing record found `%s'", newRecord, oldRecord)
		}
		if oldRecord.CID == newRecord.CID {
			return oldRecord, bosherr.New("Failed to save stemcell record `%s' (duplicate cid), existing record found `%s'", newRecord, oldRecord)
		}
	}

	records = append(records, newRecord)
	config.Stemcells = records

	err = r.configService.Save(config)
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Saving new config")
	}
	return newRecord, nil
}

func (r stemcellRepo) Find(name, version string) (StemcellRecord, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return StemcellRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		return StemcellRecord{}, false, nil
	}

	for _, oldRecord := range records {
		if oldRecord.Name == name && oldRecord.Version == version {
			return oldRecord, true, nil
		}
	}
	return StemcellRecord{}, false, nil
}

func (r stemcellRepo) FindCurrent() (StemcellRecord, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return StemcellRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	currentDiskID := config.CurrentStemcellID
	if currentDiskID == "" {
		return StemcellRecord{}, false, nil
	}

	for _, oldRecord := range config.Stemcells {
		if oldRecord.ID == currentDiskID {
			return oldRecord, true, nil
		}
	}

	return StemcellRecord{}, false, nil
}

func (r stemcellRepo) UpdateCurrent(recordID string) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	found := false
	for _, oldRecord := range config.Stemcells {
		if oldRecord.ID == recordID {
			found = true
		}
	}
	if !found {
		return bosherr.New("Verifying stemcell record exists with id `%s'", recordID)
	}

	config.CurrentStemcellID = recordID

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}
