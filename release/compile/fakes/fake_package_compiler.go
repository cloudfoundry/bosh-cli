package fakes

import (
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakePackageCompiler struct {
	CompileError    error
	CompilePackages []*bmrelease.Package
}

func NewFakePackageCompiler() *FakePackageCompiler {
	return &FakePackageCompiler{
		CompilePackages: []*bmrelease.Package{},
	}
}

func (c *FakePackageCompiler) Compile(pkg *bmrelease.Package) error {
	c.CompilePackages = append(c.CompilePackages, pkg)
	return c.CompileError
}
