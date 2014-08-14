package fakes

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeReleaseCompiler struct {
	CompileError    error
	CompilePackages []*bmrelease.Package
}

func NewFakeReleaseCompiler() *FakeReleaseCompiler {
	return &FakeReleaseCompiler{}
}

func (c *FakeReleaseCompiler) Compile(release bmrelease.Release) error {
	return c.CompileError
}
