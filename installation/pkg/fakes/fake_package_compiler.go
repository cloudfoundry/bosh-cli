package fakes

import (
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type FakePackageCompiler struct {
	CompileCompiledPackageRecord bminstallpkg.CompiledPackageRecord
	CompileError                 error
	CompilePackages              []*bmrelpkg.Package
}

func NewFakePackageCompiler() *FakePackageCompiler {
	return &FakePackageCompiler{
		CompilePackages: []*bmrelpkg.Package{},
	}
}

func (c *FakePackageCompiler) Compile(pkg *bmrelpkg.Package) (bminstallpkg.CompiledPackageRecord, error) {
	c.CompilePackages = append(c.CompilePackages, pkg)
	return c.CompileCompiledPackageRecord, c.CompileError
}
