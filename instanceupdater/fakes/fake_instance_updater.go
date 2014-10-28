package fakes

type FakeInstanceUpdater struct {
	UpdateCalled bool
	UpdateErr    error

	StartCalled bool
	StartErr    error
}

func NewFakeInstanceUpdater() *FakeInstanceUpdater {
	return &FakeInstanceUpdater{}
}

func (u *FakeInstanceUpdater) Update() error {
	u.UpdateCalled = true

	return u.UpdateErr
}

func (u *FakeInstanceUpdater) Start() error {
	u.StartCalled = true

	return u.StartErr
}
