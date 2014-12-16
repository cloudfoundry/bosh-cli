package acceptance

import (
	"fmt"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type CmdRunner interface {
	RunCommand(env map[string]string, args ...string) (string, string, int, error)
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
) CmdRunner {
	return &sshCmdRunner{
		vmUsername:     vmUsername,
		vmIP:           vmIP,
		privateKeyPath: privateKeyPath,
		runner:         boshsys.NewExecCmdRunner(logger),
	}
}

func (r *sshCmdRunner) RunCommand(env map[string]string, args ...string) (string, string, int, error) {
	exports := make([]string, len(env))
	for k, v := range env {
		exports = append(exports, fmt.Sprintf("%s=%s", k, v))
	}

	argsWithEnv := append(exports, args...)
	return r.runner.RunCommand(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", r.privateKeyPath,
		fmt.Sprintf("%s@%s", r.vmUsername, r.vmIP),
		strings.Join(argsWithEnv, " "),
	)
}
