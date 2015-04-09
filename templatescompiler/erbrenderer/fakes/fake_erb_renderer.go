package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmerbrenderer "github.com/cloudfoundry/bosh-init/templatescompiler/erbrenderer"
	bmtestutils "github.com/cloudfoundry/bosh-init/testutils"
)

type FakeERBRenderer struct {
	RenderInputs   []RenderInput
	renderBehavior map[string]renderOutput
}

type RenderInput struct {
	SrcPath string
	DstPath string
	Context bmerbrenderer.TemplateEvaluationContext
}

type renderOutput struct {
	err error
}

func NewFakeERBRender() *FakeERBRenderer {
	return &FakeERBRenderer{
		RenderInputs:   []RenderInput{},
		renderBehavior: map[string]renderOutput{},
	}
}

func (f *FakeERBRenderer) Render(srcPath, dstPath string, context bmerbrenderer.TemplateEvaluationContext) error {
	input := RenderInput{
		SrcPath: srcPath,
		DstPath: dstPath,
		Context: context,
	}
	f.RenderInputs = append(f.RenderInputs, input)
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := f.renderBehavior[inputString]

	if found {
		return output.err
	}

	return fmt.Errorf("Unsupported Input: Render('%s', '%s', '%s')", srcPath, dstPath, context)
}

func (f *FakeERBRenderer) SetRenderBehavior(srcPath, dstPath string, context bmerbrenderer.TemplateEvaluationContext, err error) error {
	input := RenderInput{
		SrcPath: srcPath,
		DstPath: dstPath,
		Context: context,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	f.renderBehavior[inputString] = renderOutput{err: err}
	return nil
}
