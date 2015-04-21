package fakes

import (
	"fmt"

	"gopkg.in/yaml.v2"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type CompileInput struct {
	Jobs                 []bireljob.Job
	DeploymentName       string
	DeploymentProperties biproperty.Map
	Stage                biui.Stage
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

func (f *FakeTemplatesCompiler) Compile(jobs []bireljob.Job, deploymentName string, deploymentProperties biproperty.Map, stage biui.Stage) error {
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

func (f *FakeTemplatesCompiler) SetCompileBehavior(jobs []bireljob.Job, deploymentName string, deploymentProperties biproperty.Map, stage biui.Stage, err error) error {
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
	bytes, err := yaml.Marshal(input)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Marshaling to string: %#v", input)
	}
	return string(bytes), nil
}
