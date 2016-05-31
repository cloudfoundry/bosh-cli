package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-init/director"
)

type VMResurrectionCmd struct {
	director boshdir.Director
}

func NewVMResurrectionCmd(director boshdir.Director) VMResurrectionCmd {
	return VMResurrectionCmd{director: director}
}

func (c VMResurrectionCmd) Run(opts VMResurrectionOpts) error {
	return c.director.EnableResurrection(bool(opts.Args.Enabled))
}
