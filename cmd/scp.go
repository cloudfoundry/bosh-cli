package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type SCPCmd struct {
	deployment  boshdir.Deployment
	uuidGen     boshuuid.Generator
	scpRunner   boshssh.SCPRunner
	ui          biui.UI
	hostBuilder boshssh.HostBuilder
}

func NewSCPCmd(
	uuidGen boshuuid.Generator,
	scpRunner boshssh.SCPRunner,
	ui biui.UI,
	hostBuilder boshssh.HostBuilder,
) SCPCmd {
	return SCPCmd{
		uuidGen:     uuidGen,
		scpRunner:   scpRunner,
		ui:          ui,
		hostBuilder: hostBuilder,
	}
}

func (c SCPCmd) Run(opts SCPOpts, deploymentFetcher boshssh.DeploymentFetcher) error {
	scpArgs := boshssh.NewSCPArgs(opts.Args.Paths, opts.Recursive)

	slug, err := scpArgs.AllOrInstanceGroupOrInstanceSlug()
	if err != nil {
		return err
	}

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	var result boshdir.SSHResult
	if opts.PrivateKey.Bytes == nil {
		c.deployment, err = deploymentFetcher()
		if err != nil {
			return err
		}

		// host key will be returned by agent over NATS
		connOpts.RawOpts = append(connOpts.RawOpts, "-o", "StrictHostKeyChecking=yes")

		result, err = c.deployment.SetUpSSH(slug, sshOpts)
		if err != nil {
			return err
		}

		defer func() {
			_ = c.deployment.CleanUpSSH(slug, sshOpts)
		}()
	} else {
		// no automatic source of host key
		connOpts.RawOpts = append(connOpts.RawOpts, "-o", "StrictHostKeyChecking=no")

		connOpts.PrivateKey = string(opts.PrivateKey.Bytes)

		host, err := c.hostBuilder.BuildHost(slug, opts.Username, deploymentFetcher)
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

	err = c.scpRunner.Run(connOpts, result, scpArgs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SCP")
	}

	return nil
}
