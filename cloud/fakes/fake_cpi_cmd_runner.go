package fakes

import (
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
)

type FakeCPICmdRunner struct {
	CurrentRunInput     []RunInput
	CurrentRunCmdOutput bicloud.CmdOutput
	CurrentRunError     error
	RunInputs           [][]RunInput
	RunCmdOutputs       []bicloud.CmdOutput
	RunErrs             []error
}

type RunInput struct {
	Context    bicloud.CmdContext
	Method     string
	Arguments  []interface{}
	ApiVersion int
}

func NewFakeCPICmdRunner() *FakeCPICmdRunner {
	return &FakeCPICmdRunner{}
}

func (r *FakeCPICmdRunner) Run(context bicloud.CmdContext, method string, apiVersion int, args ...interface{}) (bicloud.CmdOutput, error) {
	if len(r.RunCmdOutputs) > 0 {
		r.CurrentRunCmdOutput = r.RunCmdOutputs[0]

		if len(r.RunCmdOutputs) > 1 {
			r.RunCmdOutputs = r.RunCmdOutputs[1:]
		}
	}

	if len(r.RunInputs) > 0 {
		r.CurrentRunInput = r.RunInputs[0]

		if len(r.RunInputs) > 1 {
			r.RunInputs = r.RunInputs[1:]
		}
	}

	if len(r.RunErrs) > 0 {
		r.CurrentRunError = r.RunErrs[0]

		if len(r.RunErrs) > 1 {
			r.RunErrs = r.RunErrs[1:]
		}
	}

	r.CurrentRunInput = append(r.CurrentRunInput, RunInput{
		Context:    context,
		Method:     method,
		Arguments:  args,
		ApiVersion: apiVersion,
	})

	return r.CurrentRunCmdOutput, r.CurrentRunError
}
