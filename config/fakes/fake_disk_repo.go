package fakes

import (
	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
	bmconfig "github.com/cloudfoundry/bosh-init/config"
)

type FakeDiskRepo struct {
	UpdateCurrentInputs []DiskRepoUpdateCurrentInput
	updateErr           error

	findCurrentOutput diskRepoFindCurrentOutput

	SaveInputs []DiskRepoSaveInput
	saveOutput diskRepoSaveOutput

	findOutput map[string]diskRepoFindOutput

	DeleteInputs []DiskRepoDeleteInput
	DeleteErr    error

	allOutput diskRepoAllOutput
}

type DiskRepoUpdateCurrentInput struct {
	DiskID string
}

type diskRepoFindCurrentOutput struct {
	diskRecord bmconfig.DiskRecord
	found      bool
	err        error
}

type DiskRepoSaveInput struct {
	CID             string
	Size            int
	CloudProperties bmproperty.Map
}

type diskRepoSaveOutput struct {
	diskRecord bmconfig.DiskRecord
	err        error
}

type DiskRepoDeleteInput struct {
	DiskRecord bmconfig.DiskRecord
}

type diskRepoFindOutput struct {
	diskRecord bmconfig.DiskRecord
	found      bool
	err        error
}

type diskRepoAllOutput struct {
	diskRecords []bmconfig.DiskRecord
	err         error
}

func NewFakeDiskRepo() *FakeDiskRepo {
	return &FakeDiskRepo{
		UpdateCurrentInputs: []DiskRepoUpdateCurrentInput{},
		SaveInputs:          []DiskRepoSaveInput{},
		DeleteInputs:        []DiskRepoDeleteInput{},
		findOutput:          map[string]diskRepoFindOutput{},
	}
}

func (r *FakeDiskRepo) UpdateCurrent(diskID string) error {
	r.UpdateCurrentInputs = append(r.UpdateCurrentInputs, DiskRepoUpdateCurrentInput{
		DiskID: diskID,
	})
	return r.updateErr
}

func (r *FakeDiskRepo) FindCurrent() (bmconfig.DiskRecord, bool, error) {
	return r.findCurrentOutput.diskRecord, r.findCurrentOutput.found, r.findCurrentOutput.err
}

func (r *FakeDiskRepo) ClearCurrent() error {
	return nil
}

func (r *FakeDiskRepo) Save(cid string, size int, cloudProperties bmproperty.Map) (bmconfig.DiskRecord, error) {
	r.SaveInputs = append(r.SaveInputs, DiskRepoSaveInput{
		CID:             cid,
		Size:            size,
		CloudProperties: cloudProperties,
	})

	return r.saveOutput.diskRecord, r.saveOutput.err
}

func (r *FakeDiskRepo) Find(cid string) (bmconfig.DiskRecord, bool, error) {
	return r.findOutput[cid].diskRecord, r.findOutput[cid].found, r.findOutput[cid].err
}

func (r *FakeDiskRepo) All() ([]bmconfig.DiskRecord, error) {
	return r.allOutput.diskRecords, r.allOutput.err
}

func (r *FakeDiskRepo) Delete(diskRecord bmconfig.DiskRecord) error {
	r.DeleteInputs = append(r.DeleteInputs, DiskRepoDeleteInput{
		DiskRecord: diskRecord,
	})

	return r.DeleteErr
}

func (r *FakeDiskRepo) SetUpdateBehavior(err error) {
	r.updateErr = err
}

func (r *FakeDiskRepo) SetFindCurrentBehavior(diskRecord bmconfig.DiskRecord, found bool, err error) {
	r.findCurrentOutput = diskRepoFindCurrentOutput{
		diskRecord: diskRecord,
		found:      found,
		err:        err,
	}
}

func (r *FakeDiskRepo) SetSaveBehavior(diskRecord bmconfig.DiskRecord, found bool, err error) {
	r.saveOutput = diskRepoSaveOutput{
		diskRecord: diskRecord,
		err:        err,
	}
}

func (r *FakeDiskRepo) SetFindBehavior(cid string, diskRecord bmconfig.DiskRecord, found bool, err error) {
	r.findOutput[cid] = diskRepoFindOutput{
		diskRecord: diskRecord,
		found:      found,
		err:        err,
	}
}

func (r *FakeDiskRepo) SetAllBehavior(diskRecords []bmconfig.DiskRecord, err error) {
	r.allOutput = diskRepoAllOutput{
		diskRecords: diskRecords,
		err:         err,
	}
}
