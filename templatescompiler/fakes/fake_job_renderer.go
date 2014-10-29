package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeJobRenderer struct {
	RenderInputs []RenderInput
	renderErrors map[string]error
}

type RenderInput struct {
	SourcePath      string
	DestinationPath string
	Job             bmrel.Job
	Properties      map[string]interface{}
	DeploymentName  string
}

func NewFakeJobRenderer() *FakeJobRenderer {
	return &FakeJobRenderer{
		RenderInputs: []RenderInput{},
		renderErrors: map[string]error{},
	}
}

func (r *FakeJobRenderer) Render(sourcePath string, destinationPath string, job bmrel.Job, properties map[string]interface{}, deploymentName string) error {
	r.RenderInputs = append(r.RenderInputs, RenderInput{
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
		Job:             job,
		Properties:      properties,
		DeploymentName:  deploymentName,
	})
	return r.renderErrors[sourcePath]
}

func (r *FakeJobRenderer) SetRenderBehavior(sourcePath string, err error) {
	r.renderErrors[sourcePath] = err
}
