package integration_test

import (
	"crypto/tls"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/testutils"
)

var (
	originalHome string
	testHome     string
	cert         tls.Certificate
	cacertBytes  []byte
	validCACert  string
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "integration")

	BeforeEach(func() {
		originalHome = os.Getenv("HOME")

		var err error
		testHome, err = os.MkdirTemp("", "bosh-init-cli-integration")
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("HOME", testHome)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Setenv("HOME", originalHome)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(testHome)
		Expect(err).NotTo(HaveOccurred())
	})
}

var _ = BeforeSuite(func() {
	err := testutils.BuildExecutable()
	Expect(err).NotTo(HaveOccurred())
	cert, cacertBytes, err = testutils.CertSetup()
	validCACert = string(cacertBytes)
	Expect(err).ToNot(HaveOccurred())

})

func execCmd(cmdFactory cmd.Factory, args []string) {
	GinkgoHelper()
	command, err := cmdFactory.New(args)
	Expect(err).ToNot(HaveOccurred())

	err = command.Execute()
	Expect(err).ToNot(HaveOccurred())
}
