package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeReleasePackagesCompiler struct {
	CompileError   error
	CompileRelease bmrel.Release
	CompileStage   bmeventlog.Stage
}

func NewFakeReleasePackagesCompiler() *FakeReleasePackagesCompiler {
	return &FakeReleasePackagesCompiler{}
}

func (c *FakeReleasePackagesCompiler) Compile(release bmrel.Release, stage bmeventlog.Stage) error {
	c.CompileRelease = release
	c.CompileStage = stage

	return c.CompileError
}
