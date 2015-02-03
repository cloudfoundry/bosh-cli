package fakes

import (
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type FakePackageCompiler struct {
	CompileError    error
	CompilePackages []*bmrelpkg.Package
}

func NewFakePackageCompiler() *FakePackageCompiler {
	return &FakePackageCompiler{
		CompilePackages: []*bmrelpkg.Package{},
	}
}

func (c *FakePackageCompiler) Compile(pkg *bmrelpkg.Package) error {
	c.CompilePackages = append(c.CompilePackages, pkg)
	return c.CompileError
}
