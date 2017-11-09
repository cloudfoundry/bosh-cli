package ssh

import (
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type InteractiveRunner struct {
	comboRunner ComboRunner
}

func NewInteractiveRunner(comboRunner ComboRunner) InteractiveRunner {
	return InteractiveRunner{comboRunner}
}

func (r InteractiveRunner) Run(connOpts ConnectionOpts, result boshdir.SSHResult, rawCmd []string) error {
	if len(result.Hosts) == 0 {
		return bosherr.Errorf("Interactive SSH requires at least one host to be running.")
	}
	if len(result.Hosts) > 1 {
		firstHost := result.Hosts[0]
		return bosherr.Errorf("Interactive SSH only works for a single host at a time. Try `bosh ssh %s/%s`.", firstHost.Job, firstHost.IndexOrID)
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

	return r.comboRunner.Run(connOpts, result, cmdFactory)
}
