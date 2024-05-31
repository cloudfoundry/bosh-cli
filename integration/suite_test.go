package integration_test

import (
	"crypto/tls"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/testutils"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var (
	testHome string

	buildHTTPSServerCert        tls.Certificate
	buildHTTPSServerValidCACert string

	fs boshsys.FileSystem

	ui         *fakeui.FakeUI
	deps       cmd.BasicDeps
	cmdFactory cmd.Factory
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "integration")
}

var _ = BeforeSuite(func() {
	err := testutils.BuildExecutable()
	Expect(err).NotTo(HaveOccurred())

	var cacertBytes []byte
	buildHTTPSServerCert, cacertBytes, err = testutils.CertSetup()
	Expect(err).ToNot(HaveOccurred())

	buildHTTPSServerValidCACert = string(cacertBytes)
})

var _ = BeforeEach(func() {
	testHome = GinkgoT().TempDir()
	GinkgoT().Setenv("HOME", testHome)

	logger := boshlog.NewWriterLogger(boshlog.LevelNone, GinkgoWriter)
	fs = boshsys.NewOsFileSystem(logger)

	ui = &fakeui.FakeUI{}
	deps = cmd.NewBasicDepsWithFS(boshui.NewWrappingConfUI(ui, logger), fs, logger)

	cmdFactory = cmd.NewFactory(deps)
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

func createCommand(commandFactory cmd.Factory, args []string) cmd.Cmd {
	GinkgoHelper()
	command, err := commandFactory.New(args)
	Expect(err).ToNot(HaveOccurred())

	return command
}

func createAndExecCommand(commandFactory cmd.Factory, args []string) {
	GinkgoHelper()

	err := createCommand(commandFactory, args).Execute()
	Expect(err).ToNot(HaveOccurred())
}
