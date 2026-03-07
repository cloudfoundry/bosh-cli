package ssh

import (
	"context"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type InteractiveRunner struct {
	comboRunner ComboRunner
}

func NewInteractiveRunner(comboRunner ComboRunner) InteractiveRunner {
	return InteractiveRunner{comboRunner}
}

func (r InteractiveRunner) Run(connOpts ConnectionOpts, result boshdir.SSHResult, rawCmd []string) error {
	return r.RunContext(context.Background(), connOpts, result, rawCmd)
}

func (r InteractiveRunner) RunContext(ctx context.Context, connOpts ConnectionOpts, result boshdir.SSHResult, rawCmd []string) error {
	if len(result.Hosts) != 1 {
		return bosherr.Errorf("Interactive SSH only works for a single host at a time")
	}

	if len(rawCmd) != 0 {
		return bosherr.Errorf("Interactive SSH does not accept commands")
	}

	cmdFactory := func(host boshdir.Host, sshArgs SSHArgs) boshsys.Command {
		return boshsys.Command{
			Name: "ssh",
			Args: append(sshArgs.OptsForHost(host), sshArgs.LoginForHost(host)...),

			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,

			KeepAttached: true,
		}
	}

	return r.comboRunner.RunContext(ctx, connOpts, result, cmdFactory)
}
