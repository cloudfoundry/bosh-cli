package fakes

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"

	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
)

type FakeDiskRepo struct {
	// Per-VM operations
	FindCurrentForVMInputs  []string
	findCurrentForVMOutputs map[string]diskRepoFindCurrentForVMOutput

	UpdateCurrentForVMInputs []DiskRepoUpdateCurrentForVMInput
	UpdateCurrentForVMErr    error

	ClearCurrentForVMInputs []string
	ClearCurrentForVMErr    error

	// FindUnused
	findUnusedOutput diskRepoFindUnusedOutput

	// Generic operations
	SaveInputs []DiskRepoSaveInput
	saveOutput diskRepoSaveOutput

	findOutput map[string]diskRepoFindOutput

	DeleteInputs []DiskRepoDeleteInput
	DeleteErr    error

	allOutput diskRepoAllOutput
}

type DiskRepoUpdateCurrentForVMInput struct {
	VMCID  string
	DiskID string
}

type diskRepoFindCurrentForVMOutput struct {
	diskRecord biconfig.DiskRecord
	found      bool
	err        error
}

type diskRepoFindUnusedOutput struct {
	diskRecords []biconfig.DiskRecord
	err         error
}

type DiskRepoSaveInput struct {
	CID             string
	Size            int
	CloudProperties biproperty.Map
}

type diskRepoSaveOutput struct {
	diskRecord biconfig.DiskRecord
	err        error
}

type DiskRepoDeleteInput struct {
	DiskRecord biconfig.DiskRecord
}

type diskRepoFindOutput struct {
	diskRecord biconfig.DiskRecord
	found      bool
	err        error
}

type diskRepoAllOutput struct {
	diskRecords []biconfig.DiskRecord
	err         error
}

func NewFakeDiskRepo() *FakeDiskRepo {
	return &FakeDiskRepo{
		SaveInputs:              []DiskRepoSaveInput{},
		DeleteInputs:            []DiskRepoDeleteInput{},
		findOutput:              map[string]diskRepoFindOutput{},
		findCurrentForVMOutputs: map[string]diskRepoFindCurrentForVMOutput{},
	}
}

func (r *FakeDiskRepo) FindCurrentForVM(vmCID string) (biconfig.DiskRecord, bool, error) {
	r.FindCurrentForVMInputs = append(r.FindCurrentForVMInputs, vmCID)
	out := r.findCurrentForVMOutputs[vmCID]
	return out.diskRecord, out.found, out.err
}

func (r *FakeDiskRepo) SetFindCurrentForVMBehavior(vmCID string, diskRecord biconfig.DiskRecord, found bool, err error) {
	r.findCurrentForVMOutputs[vmCID] = diskRepoFindCurrentForVMOutput{
		diskRecord: diskRecord,
		found:      found,
		err:        err,
	}
}

func (r *FakeDiskRepo) UpdateCurrentForVM(vmCID string, diskID string) error {
	r.UpdateCurrentForVMInputs = append(r.UpdateCurrentForVMInputs, DiskRepoUpdateCurrentForVMInput{
		VMCID:  vmCID,
		DiskID: diskID,
	})
	return r.UpdateCurrentForVMErr
}

func (r *FakeDiskRepo) ClearCurrentForVM(vmCID string) error {
	r.ClearCurrentForVMInputs = append(r.ClearCurrentForVMInputs, vmCID)
	return r.ClearCurrentForVMErr
}

func (r *FakeDiskRepo) FindUnused() ([]biconfig.DiskRecord, error) {
	return r.findUnusedOutput.diskRecords, r.findUnusedOutput.err
}

func (r *FakeDiskRepo) SetFindUnusedBehavior(diskRecords []biconfig.DiskRecord, err error) {
	r.findUnusedOutput = diskRepoFindUnusedOutput{diskRecords: diskRecords, err: err}
}

func (r *FakeDiskRepo) Save(cid string, size int, cloudProperties biproperty.Map) (biconfig.DiskRecord, error) {
	r.SaveInputs = append(r.SaveInputs, DiskRepoSaveInput{
		CID:             cid,
		Size:            size,
		CloudProperties: cloudProperties,
	})
	return r.saveOutput.diskRecord, r.saveOutput.err
}

func (r *FakeDiskRepo) Find(cid string) (biconfig.DiskRecord, bool, error) {
	return r.findOutput[cid].diskRecord, r.findOutput[cid].found, r.findOutput[cid].err
}

func (r *FakeDiskRepo) All() ([]biconfig.DiskRecord, error) {
	return r.allOutput.diskRecords, r.allOutput.err
}

func (r *FakeDiskRepo) Delete(diskRecord biconfig.DiskRecord) error {
	r.DeleteInputs = append(r.DeleteInputs, DiskRepoDeleteInput{DiskRecord: diskRecord})
	return r.DeleteErr
}

func (r *FakeDiskRepo) SetSaveBehavior(diskRecord biconfig.DiskRecord, err error) {
	r.saveOutput = diskRepoSaveOutput{diskRecord: diskRecord, err: err}
}

func (r *FakeDiskRepo) SetFindBehavior(cid string, diskRecord biconfig.DiskRecord, found bool, err error) {
	r.findOutput[cid] = diskRepoFindOutput{diskRecord: diskRecord, found: found, err: err}
}

func (r *FakeDiskRepo) SetAllBehavior(diskRecords []biconfig.DiskRecord, err error) {
	r.allOutput = diskRepoAllOutput{diskRecords: diskRecords, err: err}
}
