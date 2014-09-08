package fakes

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type CompileInput struct {
	Jobs       []bmreljob.Job
	Deployment bmdepl.Deployment
}

type compileOutput struct {
	err error
}

type FakeTemplatesCompiler struct {
	CompileInputs []CompileInput

	compileBehavior map[string]compileOutput
}

func NewFakeTemplatesCompiler() *FakeTemplatesCompiler {
	return &FakeTemplatesCompiler{
		CompileInputs:   []CompileInput{},
		compileBehavior: map[string]compileOutput{},
	}
}

func (f *FakeTemplatesCompiler) Compile(jobs []bmreljob.Job, deployment bmdepl.Deployment) error {
	input := CompileInput{Jobs: jobs, Deployment: deployment}
	f.CompileInputs = append(f.CompileInputs, input)

	inputString, err := marshalToString(input)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling Save input")
	}
	output, found := f.compileBehavior[inputString]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: Save('%#v', '%#v')", jobs, deployment)
}

func (f *FakeTemplatesCompiler) SetCompileBehavior(jobs []bmreljob.Job, deployment bmdepl.Deployment, err error) error {
	input := CompileInput{Jobs: jobs, Deployment: deployment}
	inputString, marshalErr := marshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}
	f.compileBehavior[inputString] = compileOutput{err: err}
	return nil
}

func marshalToString(input interface{}) (string, error) {
	bytes, err := candiedyaml.Marshal(input)
	if err != nil {
		return "", bosherr.WrapError(err, "Marshaling to string: %#v", input)
	}
	return string(bytes), nil
}
