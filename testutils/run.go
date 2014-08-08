package testutils

import (
	"fmt"
	"os/exec"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
)

func BuildExecutable() error {
	session, err := RunCommand("./../bin/build")
	if session.ExitCode() != 0 {
		return fmt.Errorf("Failed to build bosh-micro:\nstdout:\n%s\nstderr:\n%s", session.Out.Contents(), session.Err.Contents())
	}
	return err
}

func RunBoshMicro(args ...string) (*gexec.Session, error) {
	return RunCommand("./../out/bosh-micro", args...)
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
