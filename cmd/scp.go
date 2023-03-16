package cmd

import (
	"errors"

	bihttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

type SCPCmd struct {
	deployment  boshdir.Deployment
	scpRunner   boshssh.SCPRunner
	hostBuilder boshssh.HostBuilder
}

func NewSCPCmd(
	scpRunner boshssh.SCPRunner,
	hostBuilder boshssh.HostBuilder,
) SCPCmd {
	return SCPCmd{
		scpRunner:   scpRunner,
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

type EnvSCPCmd struct {
	agentClientFactory bihttpagent.AgentClientFactory
	scpRunner          boshssh.SCPRunner
}

func NewEnvSCPCmd(
	agentClientFactory bihttpagent.AgentClientFactory,
	scpRunner boshssh.SCPRunner,
) EnvSCPCmd {
	return EnvSCPCmd{
		agentClientFactory: agentClientFactory,
		scpRunner:          scpRunner,
	}
}

func (c EnvSCPCmd) Run(opts SCPOpts) error {
	if opts.PrivateKey.Bytes != nil {
		return errors.New("the --private-key flag is not supported in combination with the --director flag")
	}
	if opts.Endpoint == "" || opts.Certificate == "" {
		return errors.New("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set")
	}

	agentClient, err := c.agentClientFactory.NewAgentClient("bosh-cli", opts.Endpoint, opts.Certificate)
	if err != nil {
		return err
	}

	scpArgs := boshssh.NewSCPArgs(opts.Args.Paths, opts.Recursive)

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	agentResult, err := agentClient.SetUpSSH(sshOpts.Username, sshOpts.PublicKey)
	if err != nil {
		return err
	}
	result := boshdir.SSHResult{
		Hosts: []boshdir.Host{
			{
				Username:      sshOpts.Username,
				Host:          agentResult.Ip,
				HostPublicKey: agentResult.HostPublicKey,
				Job:           "create-env-vm",
				IndexOrID:     "0",
			},
		},
	}

	defer func() {
		_, _ = agentClient.CleanUpSSH(sshOpts.Username)
	}()

	// host key will be returned by agent over HTTPS
	connOpts.RawOpts = append(connOpts.RawOpts, "-o", "StrictHostKeyChecking=yes")

	err = c.scpRunner.Run(connOpts, result, scpArgs)
	if err != nil {
		return bosherr.WrapErrorf(err, "Running SCP")
	}

	return nil
}
