package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
)

type FakeCPICmdRunner struct {
	RunInputs    []RunInput
	RunCmdOutput bmcloud.CmdOutput
	RunErr       error
}

type RunInput struct {
	Method    string
	Arguments []interface{}
}

func NewFakeCPICmdRunner() *FakeCPICmdRunner {
	return &FakeCPICmdRunner{}
}

func (r *FakeCPICmdRunner) Run(method string, args ...interface{}) (bmcloud.CmdOutput, error) {
	r.RunInputs = append(r.RunInputs, RunInput{
		Method:    method,
		Arguments: args,
	})
	return r.RunCmdOutput, r.RunErr
}
