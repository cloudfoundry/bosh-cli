package fakes

type FakeCommand struct {
	name        string
	Args        []string
	PresetError error
}

func NewFakeCommand(name string) *FakeCommand {
	return &FakeCommand{
		name: name,
		Args: []string{},
	}
}

func (f *FakeCommand) Name() string {
	return f.name
}

func (f *FakeCommand) Run(args []string) error {
	f.Args = args
	return f.PresetError
}

func (f *FakeCommand) GetArgs() []string {
	return f.Args
}
