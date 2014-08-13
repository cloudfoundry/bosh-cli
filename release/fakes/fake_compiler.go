package fakes

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeCompiler struct {
	CompileError error
}

func NewFakeCompiler() *FakeCompiler {
	return &FakeCompiler{}
}

func (c *FakeCompiler) Compile(release bmrelease.Release) error {
	return c.CompileError
}

func (c *FakeCompiler) DeterminePackageCompilationOrder(bmrelease.Release) ([]bmrelease.Package, error) {
	return []bmrelease.Package{}, nil
}
