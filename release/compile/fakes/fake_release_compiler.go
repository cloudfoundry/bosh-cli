package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeReleaseCompiler struct {
	CompileError    error
	CompilePackages []*bmrel.Package
}

func NewFakeReleaseCompiler() *FakeReleaseCompiler {
	return &FakeReleaseCompiler{}
}

func (c *FakeReleaseCompiler) Compile(release bmrel.Release) error {
	return c.CompileError
}
