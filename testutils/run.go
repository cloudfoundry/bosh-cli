package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
)

func BuildExecutable() error {
	return BuildExecutableFor("", "")
}

func BuildExecutableFor(goos, goarch string) error {
	const buildScript = "./../bin/build"

	if goos != "" {
		err := os.Setenv("GOOS", goos)
		if err != nil {
			return fmt.Errorf("failed to set GOOS=%s", goos)
		}
	}

	if goarch != "" {
		err := os.Setenv("GOARCH", goarch)
		if err != nil {
			return fmt.Errorf("failed to set GOARCH=%s", goarch)
		}
	}

	session, err := RunCommand(buildScript)
	if session.ExitCode() != 0 {
		return fmt.Errorf("failed to build bosh:\nstdout:\n%s\nstderr:\n%s", session.Out.Contents(), session.Err.Contents())
	}

	return err
}

func RunCommand(cmd string, args ...string) (*gexec.Session, error) {
	return RunComplexCommand(exec.Command(cmd, args...))
}

func RunComplexCommand(cmd *exec.Cmd) (*gexec.Session, error) {
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return nil, err
	}

	session.Wait(120 * time.Second)

	return session, nil
}
