package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type DiskRepo interface {
	// Per-VM disk tracking (replaces the old global FindCurrent/UpdateCurrent/ClearCurrent).
	FindCurrentForVM(vmCID string) (DiskRecord, bool, error)
	UpdateCurrentForVM(vmCID string, diskID string) error
	ClearCurrentForVM(vmCID string) error

	// FindUnused returns all disk records whose ID is not referenced as the
	// CurrentDiskID of any VMRecord in the deployment state.
	FindUnused() ([]DiskRecord, error)

	Save(cid string, size int, cloudProperties biproperty.Map) (DiskRecord, error)
	Find(cid string) (DiskRecord, bool, error)
	All() ([]DiskRecord, error)
	Delete(DiskRecord) error
}

type diskRepo struct {
	deploymentStateService DeploymentStateService
	uuidGenerator          boshuuid.Generator
}

func NewDiskRepo(deploymentStateService DeploymentStateService, uuidGenerator boshuuid.Generator) DiskRepo {
	return diskRepo{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
	}
}

func (r diskRepo) Save(cid string, size int, cloudProperties biproperty.Map) (DiskRecord, error) {
	config, records, err := r.load()
	if err != nil {
		return DiskRecord{}, err
	}

	oldRecord, found := r.find(records, cid)
	if found {
		return DiskRecord{}, bosherr.Errorf("Failed to save disk cid '%s', existing record found '%#v'", cid, oldRecord)
	}

	newRecord := DiskRecord{
		CID:             cid,
		Size:            size,
		CloudProperties: cloudProperties,
	}
	newRecord.ID, err = r.uuidGenerator.Generate()
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Generating disk id")
	}

	records = append(records, newRecord)
	config.Disks = records

	err = r.deploymentStateService.Save(config)
	if err != nil {
		return newRecord, bosherr.WrapError(err, "Saving new config")
	}
	return newRecord, nil
}

func (r diskRepo) FindCurrentForVM(vmCID string) (DiskRecord, bool, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return DiskRecord{}, false, bosherr.WrapError(err, "Loading existing config")
	}

	for _, vmRecord := range deploymentState.CurrentVMs {
		if vmRecord.CID == vmCID && vmRecord.CurrentDiskID != "" {
			for _, diskRecord := range deploymentState.Disks {
				if diskRecord.ID == vmRecord.CurrentDiskID {
					return diskRecord, true, nil
				}
			}
		}
	}
	return DiskRecord{}, false, nil
}

func (r diskRepo) UpdateCurrentForVM(vmCID string, diskID string) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	found := false
	for _, disk := range deploymentState.Disks {
		if disk.ID == diskID {
			found = true
			break
		}
	}
	if !found {
		return bosherr.Errorf("Verifying disk record exists with id '%s'", diskID)
	}

	for i, vmRecord := range deploymentState.CurrentVMs {
		if vmRecord.CID == vmCID {
			deploymentState.CurrentVMs[i].CurrentDiskID = diskID
			return r.deploymentStateService.Save(deploymentState)
		}
	}
	return bosherr.Errorf("VM record with CID '%s' not found", vmCID)
}

func (r diskRepo) ClearCurrentForVM(vmCID string) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	for i, vmRecord := range deploymentState.CurrentVMs {
		if vmRecord.CID == vmCID {
			deploymentState.CurrentVMs[i].CurrentDiskID = ""
			return r.deploymentStateService.Save(deploymentState)
		}
	}
	return nil
}

func (r diskRepo) FindUnused() ([]DiskRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return nil, bosherr.WrapError(err, "Loading existing config")
	}

	usedIDs := map[string]bool{}
	for _, vmRecord := range deploymentState.CurrentVMs {
		if vmRecord.CurrentDiskID != "" {
			usedIDs[vmRecord.CurrentDiskID] = true
		}
	}

	var unused []DiskRecord
	for _, disk := range deploymentState.Disks {
		if !usedIDs[disk.ID] {
			unused = append(unused, disk)
		}
	}
	return unused, nil
}

func (r diskRepo) Find(cid string) (DiskRecord, bool, error) {
	_, records, err := r.load()
	if err != nil {
		return DiskRecord{}, false, err
	}

	foundRecord, found := r.find(records, cid)
	return foundRecord, found, nil
}

func (r diskRepo) All() ([]DiskRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return []DiskRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	return deploymentState.Disks, nil
}

func (r diskRepo) Delete(diskRecord DiskRecord) error {
	config, records, err := r.load()
	if err != nil {
		return err
	}

	newRecords := []DiskRecord{}
	for _, record := range records {
		if record.ID != diskRecord.ID {
			newRecords = append(newRecords, record)
		}
	}
	config.Disks = newRecords

	// Clear the disk association from any VMRecord that references it.
	for i, vmRecord := range config.CurrentVMs {
		if vmRecord.CurrentDiskID == diskRecord.ID {
			config.CurrentVMs[i].CurrentDiskID = ""
		}
	}

	return r.deploymentStateService.Save(config)
}

func (r diskRepo) load() (DeploymentState, []DiskRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return deploymentState, []DiskRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	records := deploymentState.Disks
	if records == nil {
		return deploymentState, []DiskRecord{}, nil
	}

	return deploymentState, records, nil
}

func (r diskRepo) find(records []DiskRecord, cid string) (DiskRecord, bool) {
	for _, existingRecord := range records {
		if existingRecord.CID == cid {
			return existingRecord, true
		}
	}
	return DiskRecord{}, false
}
