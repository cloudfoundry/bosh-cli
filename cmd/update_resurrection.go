package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type UpdateResurrectionCmd struct {
	director boshdir.Director
}

func NewUpdateResurrectionCmd(director boshdir.Director) UpdateResurrectionCmd {
	return UpdateResurrectionCmd{director: director}
}

func (c UpdateResurrectionCmd) Run(opts UpdateResurrectionOpts) error {
	return c.director.EnableResurrection(bool(opts.Args.Enabled))
}
