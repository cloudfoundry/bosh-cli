package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type SSHCmd struct {
	deployment       boshdir.Deployment
	uuidGen          boshuuid.Generator
	intSSHRunner     boshssh.Runner
	nonIntSSHRunner  boshssh.Runner
	resultsSSHRunner boshssh.Runner
	ui               boshui.UI
	hostBuilder      boshssh.HostBuilder
}

func NewSSHCmd(
	uuidGen boshuuid.Generator,
	intSSHRunner boshssh.Runner,
	nonIntSSHRunner boshssh.Runner,
	resultsSSHRunner boshssh.Runner,
	ui boshui.UI,
	hostBuilder boshssh.HostBuilder,
) SSHCmd {
	return SSHCmd{
		uuidGen:          uuidGen,
		intSSHRunner:     intSSHRunner,
		nonIntSSHRunner:  nonIntSSHRunner,
		resultsSSHRunner: resultsSSHRunner,
		ui:               ui,
		hostBuilder:      hostBuilder,
	}
}

func (c SSHCmd) Run(opts SSHOpts, deploymentFetcher boshssh.DeploymentFetcher) error {
	if opts.Results || !c.ui.IsInteractive() {
		if len(opts.Command) == 0 {
			return bosherr.Errorf("Non-interactive SSH requires non-empty command")
		}
	}

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	connOpts.RawOpts = opts.RawOpts.AsStrings()

	var result boshdir.SSHResult
	if opts.PrivateKey.Bytes == nil {
		c.deployment, err = deploymentFetcher()
		if err != nil {
			return err
		}

		// host key will be returned by agent over NATS
		connOpts.RawOpts = append(connOpts.RawOpts, "-o", "StrictHostKeyChecking=yes")

		result, err = c.deployment.SetUpSSH(opts.Args.Slug, sshOpts)
		if err != nil {
			return err
		}

		defer func() {
			_ = c.deployment.CleanUpSSH(opts.Args.Slug, sshOpts)
		}()
	} else {
		// no automatic source of host key
		connOpts.RawOpts = append(connOpts.RawOpts, "-o", "StrictHostKeyChecking=no")

		connOpts.PrivateKey = string(opts.PrivateKey.Bytes)

		host, err := c.hostBuilder.BuildHost(opts.Args.Slug, opts.Username, deploymentFetcher)
		if err != nil {
			return err
		}

		result = boshdir.SSHResult{
			Hosts: []boshdir.Host{
				host,
			},
			GatewayUsername: connOpts.GatewayUsername,
			GatewayHost:     connOpts.GatewayHost,
		}
	}

	var runner boshssh.Runner

	if opts.Results {
		runner = c.resultsSSHRunner
	} else if !c.ui.IsInteractive() || len(opts.Command) > 0 {
		runner = c.nonIntSSHRunner
	} else {
		runner = c.intSSHRunner
	}

	err = runner.Run(connOpts, result, opts.Command)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SSH")
	}

	return nil
}
