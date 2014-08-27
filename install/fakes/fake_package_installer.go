package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type InstalledPackage struct {
	Package *bmrel.Package
	Target  string
}

type FakePackageInstaller struct {
	Installed []InstalledPackage
}

func (f *FakePackageInstaller) Install(pkg *bmrel.Package, targetDir string) error {
	f.Installed = append(f.Installed, InstalledPackage{
		Package: pkg,
		Target:  targetDir,
	})
	return nil
}
