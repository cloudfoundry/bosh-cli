package fakes

type FakeWorkspace struct {
	InitializeCalled bool
	InitializeError  error
}

func NewFakeWorkspace() *FakeWorkspace {
	return &FakeWorkspace{}
}

func (f *FakeWorkspace) Initialize() error {
	f.InitializeCalled = true

	return f.InitializeError
}

func (f *FakeWorkspace) BlobstorePath() string {
	return ""
}

func (f *FakeWorkspace) PackagesPath() string {
	return ""
}

func (f *FakeWorkspace) MicroBoshPath() string {
	return ""
}
