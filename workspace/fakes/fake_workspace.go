package fakes

type FakeWorkspace struct {
	InitializeCalled bool
	InitializeError  error

	LoadCalled bool
	LoadError  error
}

func NewFakeWorkspace() *FakeWorkspace {
	return &FakeWorkspace{}
}

func (f *FakeWorkspace) Initialize(manifestFile string) error {
	f.InitializeCalled = true

	return f.InitializeError
}

func (f *FakeWorkspace) Load(manifestFile string) error {
	f.LoadCalled = true

	return f.LoadError
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
