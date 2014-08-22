package fakes

type FakeWorkspace struct {
	InitializeCalled bool
	InitializeError  error
	InitializeUUID   string
}

func NewFakeWorkspace() *FakeWorkspace {
	return &FakeWorkspace{}
}

func (f *FakeWorkspace) Initialize(uuid string) error {
	f.InitializeCalled = true
	f.InitializeUUID = uuid

	return f.InitializeError
}
