package system

type ScriptCommandFactory interface {
	New(path string, args ...string) Command
	Extension() string
}

func NewScriptCommandFactory(platformName string) ScriptCommandFactory {
	if platformName == "windows" {
		return &psScriptCommandFactory{}
	}

	return &linuxScriptCommandFactory{}
}

type linuxScriptCommandFactory struct{}

func (s *linuxScriptCommandFactory) New(path string, args ...string) Command {
	return Command{
		Name: path,
		Args: args,
	}
}

func (s *linuxScriptCommandFactory) Extension() string {
	return ""
}

type psScriptCommandFactory struct{}

func (s *psScriptCommandFactory) New(path string, args ...string) Command {
	return Command{
		Name: "powershell",
		Args: append([]string{"-noprofile", "-noninteractive", path}, args...),
	}
}

func (s *psScriptCommandFactory) Extension() string {
	return ".ps1"
}
