package config

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

type DiskRecord struct {
	ID  string `json:"id"`
	CID string `json:"cid"`
}

type DiskRepo interface {
	UpdateCurrent(DiskRecord) error
	FindCurrent() (DiskRecord, bool, error)
	Save(cid string) (DiskRecord, error)
	Find(cid string) (DiskRecord, bool, error)
}

type diskRepo struct {
	configService DeploymentConfigService
	uuidGenerator boshuuid.Generator
}

func NewDiskRepo(configService DeploymentConfigService, uuidGenerator boshuuid.Generator) diskRepo {
	return diskRepo{
		configService: configService,
		uuidGenerator: uuidGenerator,
	}
}

func (r diskRepo) Save(cid string) (DiskRecord, error) {
	config, err := r.configService.Load()
	if err != nil {
		return DiskRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Disks
	if records == nil {
		records = []DiskRecord{}
	}

	oldRecord, found := r.find(records, cid)
	if found {
		return DiskRecord{}, bosherr.New("Failed to save disk cid `%s', existing record found `%s'", cid, oldRecord)
	}

	newRecord := DiskRecord{CID: cid}
	newRecord.ID, err = r.uuidGenerator.Generate()
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Generating disk id")
	}

	records = append(records, newRecord)
	config.Disks = records

	err = r.configService.Save(config)
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Saving new config")
	}
	return newRecord, nil
}

func (r diskRepo) FindCurrent() (DiskRecord, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return DiskRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	currentDiskID := config.CurrentDiskID
	if currentDiskID == "" {
		return DiskRecord{}, false, nil
	}

	for _, existingRecord := range config.Disks {
		if existingRecord.ID == currentDiskID {
			return existingRecord, true, nil
		}
	}

	return DiskRecord{}, false, nil
}

func (r diskRepo) UpdateCurrent(existingRecord DiskRecord) error {
	config, err := r.configService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	config.CurrentDiskID = existingRecord.ID

	err = r.configService.Save(config)
	if err != nil {
		return bosherr.WrapError(err, "Saving new config")
	}
	return nil
}

func (r diskRepo) Find(cid string) (DiskRecord, bool, error) {
	config, err := r.configService.Load()
	if err != nil {
		return DiskRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	records := config.Disks
	if records == nil {
		return DiskRecord{}, false, nil
	}

	foundRecord, found := r.find(records, cid)
	return foundRecord, found, nil
}

func (r diskRepo) find(records []DiskRecord, cid string) (DiskRecord, bool) {
	for _, existingRecord := range records {
		if existingRecord.CID == cid {
			return existingRecord, true
		}
	}
	return DiskRecord{}, false
}
