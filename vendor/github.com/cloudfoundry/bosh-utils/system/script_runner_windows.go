package system

func NewScriptCommand(name string, args ...string) Command {
	return Command{
		Name: "powershell",
		Args: append([]string{"-NoProfile", "-NonInteractive", name}, args...),
	}
}

func (r concreteScriptRunner) Run(script string) (string, string, error) {
	command := Command{
		Name: "powershell",
		Args: []string{"-NoProfile", "-NonInteractive", "-Command", script},
	}
	stdout, stderr, _, err := r.cmdRunner.RunComplexCommand(command)
	return stdout, stderr, err
}
