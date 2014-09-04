package fakes

import "fmt"

type FakeERBRenderer struct {
	RenderInputs   []RenderInput
	renderBehavior map[RenderInput]renderOutput
}

type RenderInput struct {
	SrcPath string
	DstPath string
}

type renderOutput struct {
	err error
}

func NewFakeERBRender() *FakeERBRenderer {
	return &FakeERBRenderer{
		RenderInputs:   []RenderInput{},
		renderBehavior: map[RenderInput]renderOutput{},
	}
}

func (f *FakeERBRenderer) Render(srcPath, dstPath string) error {
	input := RenderInput{
		SrcPath: srcPath,
		DstPath: dstPath,
	}
	f.RenderInputs = append(f.RenderInputs, input)
	output, found := f.renderBehavior[input]

	if found {
		return output.err
	}

	return fmt.Errorf("Unsupported Input: Render('%s', '%s')", srcPath, dstPath)
}

func (f *FakeERBRenderer) SetRenderBehavior(srcPath, dstPath string, err error) {
	input := RenderInput{
		SrcPath: srcPath,
		DstPath: dstPath,
	}
	f.renderBehavior[input] = renderOutput{err: err}
}
