package cmd

import boshdir "github.com/cloudfoundry/bosh-cli/director"
import . "github.com/cloudfoundry/bosh-cli/cmd/opts"

type IgnoreCmd struct {
	deployment boshdir.Deployment
}

func NewIgnoreCmd(deployment boshdir.Deployment) IgnoreCmd {
	return IgnoreCmd{deployment: deployment}
}

func (cmd IgnoreCmd) Run(opts IgnoreOpts) error {
	return cmd.deployment.Ignore(opts.Args.Slug, true)
}
