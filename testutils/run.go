package testutils

import (
	"os/exec"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
)

var BoshMicroExec string

func BuildExecutable() error {
	var err error
	BoshMicroExec, err = gexec.Build("./../../bosh-micro-cli")
	return err
}

func RunBoshMicro(args ...string) (*gexec.Session, error) {
	return RunCommand(BoshMicroExec, args...)
}

func RunCommand(cmd string, args ...string) (*gexec.Session, error) {
	command := exec.Command(cmd, args...)
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return nil, err
	}

	session.Wait()
	return session, nil
}
