package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-init/cloud"
)

type FakeCPICmdRunner struct {
	RunInputs    []RunInput
	RunCmdOutput bmcloud.CmdOutput
	RunErr       error
}

type RunInput struct {
	Context   bmcloud.CmdContext
	Method    string
	Arguments []interface{}
}

func NewFakeCPICmdRunner() *FakeCPICmdRunner {
	return &FakeCPICmdRunner{}
}

func (r *FakeCPICmdRunner) Run(context bmcloud.CmdContext, method string, args ...interface{}) (bmcloud.CmdOutput, error) {
	r.RunInputs = append(r.RunInputs, RunInput{
		Context:   context,
		Method:    method,
		Arguments: args,
	})
	return r.RunCmdOutput, r.RunErr
}
