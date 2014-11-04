package acceptance

import (
	"fmt"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type CmdRunner interface {
	RunCommand(...string) (string, string, int, error)
}

type sshCmdRunner struct {
	vmUsername     string
	vmIP           string
	privateKeyPath string
	runner         boshsys.CmdRunner
}

func NewSSHCmdRunner(
	vmUsername string,
	vmIP string,
	privateKeyPath string,
	logger boshlog.Logger,
) sshCmdRunner {
	return sshCmdRunner{
		vmUsername:     vmUsername,
		vmIP:           vmIP,
		privateKeyPath: privateKeyPath,
		runner:         boshsys.NewExecCmdRunner(logger),
	}
}

func (r sshCmdRunner) RunCommand(args ...string) (string, string, int, error) {
	argsWithEnv := append([]string{fmt.Sprintf("TMPDIR=/home/%s", r.vmUsername)}, args...)
	return r.runner.RunCommand(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", r.privateKeyPath,
		fmt.Sprintf("%s@%s", r.vmUsername, r.vmIP),
		strings.Join(argsWithEnv, " "),
	)
}
