package integration_test

import (
	"crypto/tls"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/testutils"
)

var (
	testHome string

	buildHTTPSServerCert        tls.Certificate
	buildHTTPSServerValidCACert string
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "integration")

	BeforeEach(func() {
		testHome = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", testHome)
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

func createCommand(cmdFactory cmd.Factory, args []string) cmd.Cmd {
	GinkgoHelper()
	command, err := cmdFactory.New(args)
	Expect(err).ToNot(HaveOccurred())

	return command
}

func execCmd(cmdFactory cmd.Factory, args []string) {
	GinkgoHelper()

	err := createCommand(cmdFactory, args).Execute()
	Expect(err).ToNot(HaveOccurred())
}
