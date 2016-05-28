// +build !windows

package system

import "bytes"

func NewScriptCommand(name string, args ...string) Command {
	return Command{
		Name: name,
		Args: args,
	}
}

func (r concreteScriptRunner) Run(script string) (string, string, error) {
	cmd := Command{
		Name:  "sh",
		Stdin: bytes.NewReader([]byte(script)),
	}
	stdout, stderr, _, err := r.cmdRunner.RunComplexCommand(cmd)
	return stdout, stderr, err
}
