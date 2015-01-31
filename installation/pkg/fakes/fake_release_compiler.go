package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type compileInput struct {
	Release    bmrel.Release
	Deployment bminstallmanifest.Manifest
	Stage      bmeventlog.Stage
}

type compileOutput struct {
	err error
}

type FakeReleaseCompiler struct {
	CompileInputs   []compileInput
	compileBehavior map[string]compileOutput
}

func NewFakeReleaseCompiler() *FakeReleaseCompiler {
	return &FakeReleaseCompiler{
		CompileInputs:   []compileInput{},
		compileBehavior: map[string]compileOutput{},
	}
}

func (f *FakeReleaseCompiler) Compile(release bmrel.Release, deployment bminstallmanifest.Manifest, stage bmeventlog.Stage) error {
	input := compileInput{
		Release:    release,
		Deployment: deployment,
		Stage:      stage,
	}
	f.CompileInputs = append(f.CompileInputs, input)
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := f.compileBehavior[inputString]

	if found {
		return output.err
	}

	return fmt.Errorf("Unsupported Compile Input: %s\nAvailable inputs: %s", inputString, f.compileBehavior)
}

func (f *FakeReleaseCompiler) SetCompileBehavior(release bmrel.Release, deployment bminstallmanifest.Manifest, stage bmeventlog.Stage, err error) error {
	input := compileInput{
		Release:    release,
		Deployment: deployment,
		Stage:      stage,
	}
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	f.compileBehavior[inputString] = compileOutput{err: err}
	return nil
}
