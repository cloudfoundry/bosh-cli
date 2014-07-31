package test_helpers

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var BoshMicroExec string

func GetFilePath(input_dir string, fileName string) string {
	return filepath.Join(os.Getenv("PWD"), input_dir, fileName)
}

func RemoveAllFiles(args ...string) {
	for _, arg := range args {
		err := os.Remove(arg)
		Expect(err).NotTo(HaveOccurred())
	}
}

func BuildExecutable() {
	var err error
	BoshMicroExec, err = gexec.Build("./../../bosh-micro-cli")
	Expect(err).NotTo(HaveOccurred())
}

func RunBoshMicro(args ...string) *gexec.Session {
	session := RunCommand(BoshMicroExec, args...)
	return session
}

func RunCommand(cmd string, args ...string) *gexec.Session {
	command := exec.Command(cmd, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait()
	return session
}

func StubBoshMicroPath() {
	oldHome := os.Getenv("HOME")
	boshMicroPath, err := ioutil.TempDir("", "micro-bosh-cli-integration")
	Expect(err).NotTo(HaveOccurred())
	BeforeEach(func() {
		os.Setenv("HOME", boshMicroPath)
	})
	AfterEach(func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(boshMicroPath)
	})
}
