package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeReleaseCompiler struct {
	CompileError   error
	CompileRelease bmrel.Release
}

func NewFakeReleaseCompiler() *FakeReleaseCompiler {
	return &FakeReleaseCompiler{}
}

func (c *FakeReleaseCompiler) Compile(release bmrel.Release) error {
	c.CompileRelease = release

	return c.CompileError
}
