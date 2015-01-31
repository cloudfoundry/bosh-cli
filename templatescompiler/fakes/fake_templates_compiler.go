package fakes

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type CompileInput struct {
	Jobs                 []bmrel.Job
	DeploymentName       string
	DeploymentProperties map[string]interface{}
	Stage                bmeventlog.Stage
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

func (f *FakeTemplatesCompiler) Compile(jobs []bmrel.Job, deploymentName string, deploymentProperties map[string]interface{}, stage bmeventlog.Stage) error {
	input := CompileInput{
		Jobs:                 jobs,
		DeploymentName:       deploymentName,
		DeploymentProperties: deploymentProperties,
		Stage:                stage,
	}
	f.CompileInputs = append(f.CompileInputs, input)

	inputString, err := marshalToString(input)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling Save input")
	}
	output, found := f.compileBehavior[inputString]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: Save('%#v', '%#v', '%#v')", jobs, deploymentName, deploymentProperties)
}

func (f *FakeTemplatesCompiler) SetCompileBehavior(jobs []bmrel.Job, deploymentName string, deploymentProperties map[string]interface{}, stage bmeventlog.Stage, err error) error {
	input := CompileInput{
		Jobs:                 jobs,
		DeploymentName:       deploymentName,
		DeploymentProperties: deploymentProperties,
		Stage:                stage,
	}
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
		return "", bosherr.WrapErrorf(err, "Marshaling to string: %#v", input)
	}
	return string(bytes), nil
}
