package acceptance_test

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"

	. "github.com/cloudfoundry/bosh-micro-cli/acceptance"
)

var _ = Describe("bosh-micro", func() {
	var (
		sshCmdRunner CmdRunner
		testEnv      Environment
	)

	BeforeSuite(func() {
		localEnv, err := parseEnv()
		Expect(err).NotTo(HaveOccurred())

		logger := boshlog.NewLogger(boshlog.LevelDebug)
		fileSystem := boshsys.NewOsFileSystem(logger)
		testEnv = NewRemoteTestEnvironment(
			localEnv.vmUsername,
			localEnv.vmIP,
			localEnv.privateKeyPath,
			fileSystem,
			logger,
		)

		sshCmdRunner = NewSSHCmdRunner(
			localEnv.vmUsername,
			localEnv.vmIP,
			localEnv.privateKeyPath,
			logger,
		)

		err = bmtestutils.BuildExecutableForArch("linux-amd64")
		Expect(err).NotTo(HaveOccurred())

		boshMicroPath := "./../out/bosh-micro"
		Expect(fileSystem.FileExists(boshMicroPath)).To(BeTrue())
		err = testEnv.Copy("bosh-micro", boshMicroPath)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("stemcell", localEnv.stemcellURL)
		Expect(err).NotTo(HaveOccurred())
		err = testEnv.DownloadOrCopy("cpiRelease", localEnv.cpiReleaseURL)
		Expect(err).NotTo(HaveOccurred())
	})

	ItSetsSSHPassword := func(username, password, hostname string) {
		authMethods := []ssh.AuthMethod{
			ssh.Password(password),
		}

		sshConfig := &ssh.ClientConfig{
			User: username,
			Auth: authMethods,
		}

		conn, _ := ssh.Dial("tcp", hostname+":22", sshConfig)
		session, _ := conn.NewSession()
		defer session.Close()

		var stdoutBuf bytes.Buffer
		session.Stdout = &stdoutBuf
		session.Run("echo ok")
		Expect(stdoutBuf.String()).To(ContainSubstring("ok"))
	}

	It("is able to deploy a CPI release with a stemcell", func() {
		manifestPath := "./manifest.yml"
		manifestContents, err := ioutil.ReadFile(manifestPath)
		Expect(err).ToNot(HaveOccurred())
		testEnv.WriteContent("manifest", manifestContents)

		_, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deployment", testEnv.Path("manifest"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		stdout, _, exitCode, err := sshCmdRunner.RunCommand(testEnv.Path("bosh-micro"), "deploy", testEnv.Path("cpiRelease"), testEnv.Path("stemcell"))
		Expect(err).ToNot(HaveOccurred())
		Expect(exitCode).To(Equal(0))
		Expect(stdout).To(ContainSubstring("uploading stemcell"))
		Expect(stdout).To(ContainSubstring("Creating VM from"))
		Expect(stdout).To(ContainSubstring("Waiting for the agent"))
		Expect(stdout).To(ContainSubstring("Applying micro BOSH spec"))
		Expect(stdout).To(ContainSubstring("Starting agent services"))
		Expect(stdout).To(ContainSubstring("Waiting for the director"))

		ItSetsSSHPassword("vcap", "sshpassword", "10.244.0.42")
	})
})

type localEnvironment struct {
	vmUsername     string
	vmIP           string
	privateKeyPath string
	stemcellURL    string
	cpiReleaseURL  string
}

func parseEnv() (localEnvironment, error) {
	env := localEnvironment{
		vmUsername:     os.Getenv("BOSH_MICRO_VM_USERNAME"),
		vmIP:           os.Getenv("BOSH_MICRO_VM_IP"),
		privateKeyPath: os.Getenv("BOSH_MICRO_PRIVATE_KEY"),
		stemcellURL:    os.Getenv("BOSH_MICRO_STEMCELL"),
		cpiReleaseURL:  os.Getenv("BOSH_MICRO_CPI_RELEASE"),
	}

	var err error
	if env.vmUsername == "" {
		fmt.Println("BOSH_MICRO_VM_USERNAME must be set")
		err = errors.New("")
	}
	if env.vmIP == "" {
		fmt.Println("BOSH_MICRO_VM_IP must be set")
		err = errors.New("")
	}
	if env.privateKeyPath == "" {
		fmt.Println("BOSH_MICRO_PRIVATE_KEY must be set")
		err = errors.New("")
	}
	if env.stemcellURL == "" {
		fmt.Println("BOSH_MICRO_STEMCELL must be set")
		err = errors.New("")
	}
	if env.cpiReleaseURL == "" {
		fmt.Println("BOSH_MICRO_CPI_RELEASE must be set")
		err = errors.New("")
	}

	return env, err
}
