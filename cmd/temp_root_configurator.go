package cmd

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type TempRootConfigurator interface {
	PrepareAndSetTempRoot(path string) error
}

type tempRootConfigurator struct {
	fs boshsys.FileSystem
}

func NewTempRootConfigurator(fs boshsys.FileSystem) TempRootConfigurator {
	return &tempRootConfigurator{fs: fs}
}

func (c *tempRootConfigurator) PrepareAndSetTempRoot(path string) error {
	err := c.fs.ChangeTempRoot(path)
	if err != nil {
		return err
	}
	return nil
}
