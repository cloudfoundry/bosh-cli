package fakes

type FakeInstanceUpdater struct {
	UpdateCalled bool
	UpdateErr    error
}

func NewFakeInstanceUpdater() *FakeInstanceUpdater {
	return &FakeInstanceUpdater{}
}

func (u *FakeInstanceUpdater) Update() error {
	u.UpdateCalled = true

	return u.UpdateErr
}
