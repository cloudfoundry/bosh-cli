package config

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type VMRepo interface {
	FindAll() ([]VMRecord, error)
	// Save records a newly-created VM. If a pending record (CID == "") with the
	// same JobName+InstanceID already exists (left by a previous deletion that
	// preserved disk association), it is reused; otherwise a new record is appended.
	Save(jobName string, instanceID int, cid string, staticIP string) (VMRecord, error)
	UpdateCurrentDisk(vmCID string, diskID string) error
	// Delete removes the VMRecord for vmCID.  When the record holds a
	// CurrentDiskID the disk association is preserved: the record's CID and
	// MbusURL are cleared but the record itself remains so that the next
	// deployment of the same instance can re-attach its disk.
	Delete(vmCID string) error
	ClearAll() error
}

type vMRepo struct {
	deploymentStateService DeploymentStateService
	uuidGenerator          boshuuid.Generator
}

func NewVMRepo(deploymentStateService DeploymentStateService, uuidGenerator boshuuid.Generator) VMRepo {
	return vMRepo{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
	}
}

func (r vMRepo) FindAll() ([]VMRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return nil, bosherr.WrapError(err, "Loading existing config")
	}

	var active []VMRecord
	for _, rec := range deploymentState.CurrentVMs {
		if rec.CID != "" {
			active = append(active, rec)
		}
	}
	return active, nil
}

func (r vMRepo) Save(jobName string, instanceID int, cid string, staticIP string) (VMRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return VMRecord{}, bosherr.WrapError(err, "Loading existing config")
	}

	// Reuse a pending record for the same instance (left by Delete with a disk).
	for i, rec := range deploymentState.CurrentVMs {
		if rec.JobName == jobName && rec.InstanceID == instanceID && rec.CID == "" {
			deploymentState.CurrentVMs[i].CID = cid
			deploymentState.CurrentVMs[i].StaticIP = staticIP
			if err = r.deploymentStateService.Save(deploymentState); err != nil {
				return VMRecord{}, bosherr.WrapError(err, "Saving new config")
			}
			return deploymentState.CurrentVMs[i], nil
		}
	}

	// Create a fresh record.
	id, err := r.uuidGenerator.Generate()
	if err != nil {
		return VMRecord{}, bosherr.WrapError(err, "Generating VM record ID")
	}
	record := VMRecord{
		ID:         id,
		JobName:    jobName,
		InstanceID: instanceID,
		CID:        cid,
		StaticIP:   staticIP,
	}
	deploymentState.CurrentVMs = append(deploymentState.CurrentVMs, record)
	if err = r.deploymentStateService.Save(deploymentState); err != nil {
		return VMRecord{}, bosherr.WrapError(err, "Saving new config")
	}
	return record, nil
}

func (r vMRepo) UpdateCurrentDisk(vmCID string, diskID string) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	for i, rec := range deploymentState.CurrentVMs {
		if rec.CID == vmCID {
			deploymentState.CurrentVMs[i].CurrentDiskID = diskID
			return r.deploymentStateService.Save(deploymentState)
		}
	}
	return bosherr.Errorf("VM record with CID '%s' not found", vmCID)
}

func (r vMRepo) Delete(vmCID string) error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	for i, rec := range deploymentState.CurrentVMs {
		if rec.CID == vmCID {
			if rec.CurrentDiskID != "" {
				// Keep the record so the disk association survives recreation.
				deploymentState.CurrentVMs[i].CID = ""
				deploymentState.CurrentVMs[i].StaticIP = ""
			} else {
				deploymentState.CurrentVMs = append(
					deploymentState.CurrentVMs[:i],
					deploymentState.CurrentVMs[i+1:]...,
				)
			}
			return r.deploymentStateService.Save(deploymentState)
		}
	}
	return nil
}

func (r vMRepo) ClearAll() error {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}
	deploymentState.CurrentVMs = nil
	return r.deploymentStateService.Save(deploymentState)
}
