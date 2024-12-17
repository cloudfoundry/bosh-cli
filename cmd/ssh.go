package cmd

import (
	"errors"

	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

type SSHCmd struct {
	deployment       boshdir.Deployment
	intSSHRunner     boshssh.Runner
	nonIntSSHRunner  boshssh.Runner
	resultsSSHRunner boshssh.Runner
	ui               boshui.UI
	hostBuilder      boshssh.HostBuilder
}

func NewSSHCmd(
	intSSHRunner boshssh.Runner,
	nonIntSSHRunner boshssh.Runner,
	resultsSSHRunner boshssh.Runner,
	ui boshui.UI,
	hostBuilder boshssh.HostBuilder,
) SSHCmd {
	return SSHCmd{
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

type EnvSSHCmd struct {
	agentClientFactory bihttpagent.AgentClientFactory
	intSSHRunner       boshssh.Runner
	nonIntSSHRunner    boshssh.Runner
	resultsSSHRunner   boshssh.Runner
	ui                 boshui.UI
}

func NewEnvSSHCmd(
	agentClientFactory bihttpagent.AgentClientFactory,
	intSSHRunner boshssh.Runner,
	nonIntSSHRunner boshssh.Runner,
	resultsSSHRunner boshssh.Runner,
	ui boshui.UI,
) EnvSSHCmd {
	return EnvSSHCmd{
		agentClientFactory: agentClientFactory,
		intSSHRunner:       intSSHRunner,
		nonIntSSHRunner:    nonIntSSHRunner,
		resultsSSHRunner:   resultsSSHRunner,
		ui:                 ui,
	}
}

func (c EnvSSHCmd) Run(opts SSHOpts) error {
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

	sshOpts, connOpts, err := opts.GatewayFlags.AsSSHOpts()
	if err != nil {
		return err
	}

	connOpts.RawOpts = opts.RawOpts.AsStrings()
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
