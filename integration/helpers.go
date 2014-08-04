package integration

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

func GetFilePath(inputDir string, fileName string) string {
	return filepath.Join(os.Getenv("PWD"), inputDir, fileName)
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

func GenerateDeploymentManifest() string {
	tmpFile, err := ioutil.TempFile("", "bosh-micro-cli-deployment-manifest")
	Expect(err).NotTo(HaveOccurred())
	deploymentManifestFilePath := tmpFile.Name()

	err = ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	return deploymentManifestFilePath
}

func GenerateCPIRelease() string {
	tmpFile, err := ioutil.TempFile("", "bosh-micro-cli-cpi-release")
	Expect(err).NotTo(HaveOccurred())
	cpiFilePath := tmpFile.Name()

	err = ioutil.WriteFile(cpiFilePath, []byte(""), os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	return cpiFilePath
}
