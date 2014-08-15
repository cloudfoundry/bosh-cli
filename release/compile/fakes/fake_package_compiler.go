package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakePackageCompiler struct {
	CompileError    error
	CompilePackages []*bmrel.Package
}

func NewFakePackageCompiler() *FakePackageCompiler {
	return &FakePackageCompiler{
		CompilePackages: []*bmrel.Package{},
	}
}

func (c *FakePackageCompiler) Compile(pkg *bmrel.Package) error {
	c.CompilePackages = append(c.CompilePackages, pkg)
	return c.CompileError
}
