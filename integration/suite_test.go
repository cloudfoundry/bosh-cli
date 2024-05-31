package integration_test

import (
	"crypto/tls"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/testutils"
)

var (
	originalHome string
	testHome     string

	buildHTTPSServerCert        tls.Certificate
	buildHTTPSServerValidCACert string
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

	var cacertBytes []byte
	buildHTTPSServerCert, cacertBytes, err = testutils.CertSetup()
	Expect(err).ToNot(HaveOccurred())

	buildHTTPSServerValidCACert = string(cacertBytes)
})

func buildHTTPSServer() (string, *ghttp.Server) {
	GinkgoHelper()

	server := ghttp.NewUnstartedServer()
	server.HTTPTestServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{buildHTTPSServerCert},
	}

	server.HTTPTestServer.StartTLS()

	return buildHTTPSServerValidCACert, server
}

func execCmd(cmdFactory cmd.Factory, args []string) {
	GinkgoHelper()
	command, err := cmdFactory.New(args)
	Expect(err).ToNot(HaveOccurred())

	err = command.Execute()
	Expect(err).ToNot(HaveOccurred())
}
