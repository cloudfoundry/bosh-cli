package fakes

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type CompileInput struct {
	Jobs                 []bmreljob.Job
	DeploymentName       string
	DeploymentProperties bmproperty.Map
	Stage                bmui.Stage
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

func (f *FakeTemplatesCompiler) Compile(jobs []bmreljob.Job, deploymentName string, deploymentProperties bmproperty.Map, stage bmui.Stage) error {
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

func (f *FakeTemplatesCompiler) SetCompileBehavior(jobs []bmreljob.Job, deploymentName string, deploymentProperties bmproperty.Map, stage bmui.Stage, err error) error {
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
