package erbrenderer

type ERBRenderer interface {
	Render(srcPath, dstPath string) error
}
