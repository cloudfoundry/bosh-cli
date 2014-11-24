package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

// StemcellRepo persists stemcells metadata
type StemcellRepo interface {
	UpdateCurrent(recordID string) error
	FindCurrent() (StemcellRecord, bool, error)
	ClearCurrent() error
	Save(name, version, cid string) (StemcellRecord, error)
	Find(name, version string) (StemcellRecord, bool, error)
	All() ([]StemcellRecord, error)
	Delete(StemcellRecord) error
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
	stemcellRecord := StemcellRecord{}

	err := r.updateConfig(func(config *DeploymentFile) error {
		records := config.Stemcells
		if records == nil {
			records = []StemcellRecord{}
		}

		newRecord := StemcellRecord{
			Name:    name,
			Version: version,
			CID:     cid,
		}
		var err error
		newRecord.ID, err = r.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating stemcell id")
		}

		for _, oldRecord := range records {
			if oldRecord.Name == newRecord.Name && oldRecord.Version == newRecord.Version {
				return bosherr.New("Failed to save stemcell record `%s' (duplicate name/version), existing record found `%s'", newRecord, oldRecord)
			}
			if oldRecord.CID == newRecord.CID {
				return bosherr.New("Failed to save stemcell record `%s' (duplicate cid), existing record found `%s'", newRecord, oldRecord)
			}
		}

		records = append(records, newRecord)
		config.Stemcells = records

		stemcellRecord = newRecord

		return nil
	})

	return stemcellRecord, err
}

func (r stemcellRepo) Find(name, version string) (StemcellRecord, bool, error) {
	_, records, err := r.load()
	if err != nil {
		return StemcellRecord{}, false, err
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

func (r stemcellRepo) All() ([]StemcellRecord, error) {
	config, err := r.configService.Load()
	if err != nil {
		return []StemcellRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	return config.Stemcells, nil
}

func (r stemcellRepo) Delete(stemcellRecord StemcellRecord) error {
	config, records, err := r.load()
	if err != nil {
		return err
	}

	newRecords := []StemcellRecord{}
	for _, record := range records {
		if stemcellRecord.ID != record.ID {
			newRecords = append(newRecords, record)
		}
	}

	config.Stemcells = newRecords

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving config")
	}

	return nil
}

func (r stemcellRepo) UpdateCurrent(recordID string) error {
	return r.updateConfig(func(config *DeploymentFile) error {
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

		return nil
	})
}

func (r stemcellRepo) ClearCurrent() error {
	return r.updateConfig(func(config *DeploymentFile) error {
		config.CurrentStemcellID = ""

		return nil
	})
}

func (r stemcellRepo) updateConfig(updateFunc func(*DeploymentFile) error) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	err = updateFunc(&config)
	if err != nil {
		return err
	}

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}

	return nil
}

func (r stemcellRepo) load() (DeploymentFile, []StemcellRecord, error) {
	config, err := r.configService.Load()
	if err != nil {
		return config, []StemcellRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Stemcells
	if records == nil {
		return config, []StemcellRecord{}, nil
	}

	return config, records, nil
}
