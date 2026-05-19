package fakes

import (
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
)

type FakeVMRepo struct {
	SaveInputs  []VMRepoSaveInput
	SaveOutput  VMRepoSaveOutput
	SaveErr     error

	UpdateCurrentDiskInputs []VMRepoUpdateCurrentDiskInput
	UpdateCurrentDiskErr    error

	DeleteInputs []string
	DeleteErr    error

	ClearAllCalled bool
	ClearAllErr    error

	findAllOutput vmRepoFindAllOutput
}

type VMRepoSaveInput struct {
	JobName    string
	InstanceID int
	CID        string
	StaticIP   string
}

type VMRepoSaveOutput struct {
	Record biconfig.VMRecord
}

type VMRepoUpdateCurrentDiskInput struct {
	VMCID  string
	DiskID string
}

type vmRepoFindAllOutput struct {
	records []biconfig.VMRecord
	err     error
}

func NewFakeVMRepo() *FakeVMRepo {
	return &FakeVMRepo{}
}

func (r *FakeVMRepo) FindAll() ([]biconfig.VMRecord, error) {
	return r.findAllOutput.records, r.findAllOutput.err
}

func (r *FakeVMRepo) SetFindAllBehavior(records []biconfig.VMRecord, err error) {
	r.findAllOutput = vmRepoFindAllOutput{records: records, err: err}
}

func (r *FakeVMRepo) Save(jobName string, instanceID int, cid string, staticIP string) (biconfig.VMRecord, error) {
	r.SaveInputs = append(r.SaveInputs, VMRepoSaveInput{
		JobName:    jobName,
		InstanceID: instanceID,
		CID:        cid,
		StaticIP:   staticIP,
	})
	return r.SaveOutput.Record, r.SaveErr
}

func (r *FakeVMRepo) UpdateCurrentDisk(vmCID string, diskID string) error {
	r.UpdateCurrentDiskInputs = append(r.UpdateCurrentDiskInputs, VMRepoUpdateCurrentDiskInput{
		VMCID:  vmCID,
		DiskID: diskID,
	})
	return r.UpdateCurrentDiskErr
}

func (r *FakeVMRepo) Delete(vmCID string) error {
	r.DeleteInputs = append(r.DeleteInputs, vmCID)
	return r.DeleteErr
}

func (r *FakeVMRepo) ClearAll() error {
	r.ClearAllCalled = true
	return r.ClearAllErr
}
