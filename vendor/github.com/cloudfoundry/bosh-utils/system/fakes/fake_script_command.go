package fakes

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FakeCommandFactory struct {
	ReturnExtension string
}

func (f *FakeCommandFactory) New(path string, args ...string) boshsys.Command {
	return boshsys.Command{
		Name: path,
		Args: args,
	}
}

func (f *FakeCommandFactory) Extension() string {
	return f.ReturnExtension
}
