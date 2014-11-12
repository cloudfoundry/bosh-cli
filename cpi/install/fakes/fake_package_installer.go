package fakes

import (
	"fmt"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type InstallInput struct {
	Package *bmrel.Package
	Target  string
}
type installOutput struct {
	err error
}

type FakePackageInstaller struct {
	InstallInputs   []InstallInput
	installBehavior map[InstallInput]installOutput
}

func NewFakePackageInstaller() *FakePackageInstaller {
	return &FakePackageInstaller{
		InstallInputs:   []InstallInput{},
		installBehavior: map[InstallInput]installOutput{},
	}
}

func (f *FakePackageInstaller) Install(pkg *bmrel.Package, targetDir string) error {
	input := InstallInput{Package: pkg, Target: targetDir}
	f.InstallInputs = append(f.InstallInputs, input)
	output, found := f.installBehavior[input]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: Install(%#v, '%s')", pkg, targetDir)
}

func (f *FakePackageInstaller) SetInstallBehavior(pkg *bmrel.Package, targetDir string, err error) {
	f.installBehavior[InstallInput{Package: pkg, Target: targetDir}] = installOutput{err: err}
}
