package fakes

type FakeVMRepo struct {
	UpdateCurrentCID string
	UpdateCurrentErr error

	ClearCurrentCalled bool
	ClearCurrentErr    error

	findCurrentOutput   vmRepoFindCurrentOutput
	findCurrentIpOutput vmRepoFindCurrentIpOutput

	updateCurrentIp    string
	updateCurrentIpErr error
}

type vmRepoFindCurrentOutput struct {
	cid   string
	found bool
	err   error
}

type vmRepoFindCurrentIpOutput struct {
	ip    string
	found bool
	err   error
}

func NewFakeVMRepo() *FakeVMRepo {
	return &FakeVMRepo{}
}

func (r *FakeVMRepo) FindCurrent() (cid string, found bool, err error) {
	return r.findCurrentOutput.cid, r.findCurrentOutput.found, r.findCurrentOutput.err
}

func (r *FakeVMRepo) SetFindCurrentBehavior(cid string, found bool, err error) {
	r.findCurrentOutput = vmRepoFindCurrentOutput{
		cid:   cid,
		found: found,
		err:   err,
	}
}

func (r *FakeVMRepo) UpdateCurrent(cid string) error {
	r.UpdateCurrentCID = cid
	return r.UpdateCurrentErr
}

func (r *FakeVMRepo) ClearCurrent() error {
	r.ClearCurrentCalled = true
	return r.ClearCurrentErr
}

func (r *FakeVMRepo) FindCurrentIP() (string, bool, error) {
	return r.findCurrentIpOutput.ip, r.findCurrentIpOutput.found, r.findCurrentIpOutput.err
}

func (r *FakeVMRepo) UpdateCurrentIP(ip string) error {
	r.updateCurrentIp = ip
	return r.updateCurrentIpErr
}

func (r *FakeVMRepo) SetFindCurrentIPBehavior(ip string, found bool, err error) {
	r.findCurrentIpOutput = vmRepoFindCurrentIpOutput{
		ip:    ip,
		found: found,
		err:   err,
	}
}
