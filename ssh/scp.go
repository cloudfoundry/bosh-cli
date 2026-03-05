package ssh

import (
	"context"

	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type SCPRunnerImpl struct {
	comboRunner ComboRunner
}

func NewSCPRunner(comboRunner ComboRunner) SCPRunnerImpl {
	return SCPRunnerImpl{comboRunner}
}

func (r SCPRunnerImpl) Run(connOpts ConnectionOpts, result boshdir.SSHResult, scpArgs SCPArgs) error {
	return r.RunContext(context.Background(), connOpts, result, scpArgs)
}

func (r SCPRunnerImpl) RunContext(ctx context.Context, connOpts ConnectionOpts, result boshdir.SSHResult, scpArgs SCPArgs) error {
	cmdFactory := func(host boshdir.Host, sshArgs SSHArgs) boshsys.Command {
		return boshsys.Command{
			Name: "scp",
			Args: append(sshArgs.OptsForHost(host), scpArgs.ForHost(host)...), // src/dst args come last
		}
	}

	return r.comboRunner.RunContext(ctx, connOpts, result, cmdFactory)
}
