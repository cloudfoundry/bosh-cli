package test_helpers

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var BoshMicroExec string

func GetFilePath(input_dir string, fileName string) string {
	return filepath.Join(os.Getenv("PWD"), input_dir, fileName)
}

func RemoveAllFiles(args ...string) {
	for _, arg := range args {
		os.Remove(arg)
	}
}

func BuildExecutable() {
	var err error
	BoshMicroExec, err = gexec.Build("./../../bosh-micro-cli")
	Ω(err).ShouldNot(HaveOccurred())
}

func RunBoshMicro(args ...string) *Session {
	session := RunCommand(BoshMicroExec, args...)
	return session
}

func RunCommand(cmd string, args ...string) *Session {
	command := exec.Command(cmd, args...)
	session, err := Start(command, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	session.Wait()
	return session
}
