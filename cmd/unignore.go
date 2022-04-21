package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type UnignoreCmd struct {
	deployment boshdir.Deployment
}

func NewUnignoreCmd(deployment boshdir.Deployment) UnignoreCmd {
	return UnignoreCmd{deployment: deployment}
}

func (cmd UnignoreCmd) Run(opts UnignoreOpts) error {
	return cmd.deployment.Ignore(opts.Args.Slug, false)
}
