package acceptance

import (
	"fmt"
	"os"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type MicroSSH interface {
	RunCommand(cmd string) (stdout, stderr string, exitCode int, err error)
	RunCommandWithSudo(cmd string) (stdout, stderr string, exitCode int, err error)
}

type microSSH struct {
	vmUsername     string
	vmIP           string
	privateKeyPath string
	microUsername  string
	microIP        string
	microPassword  string
	runner         boshsys.CmdRunner
	fileSystem     boshsys.FileSystem
}

func NewMicroSSH(
  vmUsername string,
	vmIP string,
	privateKeyPath string,
	microUsername string,
	microIP string,
	microPassword string,
	fileSystem boshsys.FileSystem,
	logger boshlog.Logger,
) microSSH {
	return microSSH{
		vmUsername:     vmUsername,
		vmIP:           vmIP,
		privateKeyPath: privateKeyPath,
		microUsername: microUsername,
		microIP: microIP,
		microPassword: microPassword,
		runner:         boshsys.NewExecCmdRunner(logger),
		fileSystem:     fileSystem,
	}
}

func (s microSSH) setupSSH() (*os.File, error) {
	sshConfigFile, err := s.fileSystem.TempFile("ssh-config")
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating temp ssh-config file")
	}

	success := false
	defer func() {
		if !success {
			s.fileSystem.RemoveAll(sshConfigFile.Name())
		}
	}()

	sshConfigTemplate := `
Host vagrant-vm
	HostName %s
	User %s
	Port 22
	StrictHostKeyChecking no
	IdentityFile %s
Host warden-vm
	Hostname %s
	User %s
	StrictHostKeyChecking no
	ProxyCommand ssh -F %s vagrant-vm netcat -w 120 %%h %%p
`
	sshConfig := fmt.Sprintf(
		sshConfigTemplate,
		s.vmIP,
		s.vmUsername,
	  s.privateKeyPath,
		s.microIP,
		s.microUsername,
		sshConfigFile.Name(),
	)

	err = s.fileSystem.WriteFileString(sshConfigFile.Name(), sshConfig)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Writing to temp ssh-config file: '%s'", sshConfigFile.Name())
	}

	success = true
	return sshConfigFile, nil
}

func (s microSSH) RunCommand(cmd string) (stdout, stderr string, exitCode int, err error) {
	sshConfigFile, err := s.setupSSH()
	if err != nil {
		return "", "", -1, bosherr.WrapError(err, "Setting up SSH")
	}
	defer s.fileSystem.RemoveAll(sshConfigFile.Name())

	return s.runner.RunCommand(
		"sshpass",
		"-p"+s.microPassword,
		"ssh",
		"warden-vm",
		"-F",
		sshConfigFile.Name(),
		cmd,
	)
}

func (s microSSH) RunCommandWithSudo(cmd string) (stdout, stderr string, exitCode int, err error) {
	sshConfigFile, err := s.setupSSH()
	if err != nil {
		return "", "", -1, bosherr.WrapError(err, "Setting up SSH")
	}
	defer s.fileSystem.RemoveAll(sshConfigFile.Name())

	return s.runner.RunCommand(
		"sshpass",
			"-p"+s.microPassword,
		"ssh",
		"warden-vm",
		"-F",
		sshConfigFile.Name(),
		fmt.Sprintf("echo %s | sudo -p '' -S %s", s.microPassword, cmd),
	)
}
