package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeReleasePackagesCompiler struct {
	CompileError   error
	CompileRelease bmrel.Release
}

func NewFakeReleasePackagesCompiler() *FakeReleasePackagesCompiler {
	return &FakeReleasePackagesCompiler{}
}

func (c *FakeReleasePackagesCompiler) Compile(release bmrel.Release) error {
	c.CompileRelease = release

	return c.CompileError
}
