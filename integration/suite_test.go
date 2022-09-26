package integration_test

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/testutils"
	bitestutils "github.com/cloudfoundry/bosh-cli/v7/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	RunSpecs(t, "integration")
}

var _ = BeforeSuite(func() {
	err := bitestutils.BuildExecutable()
	Expect(err).NotTo(HaveOccurred())
	cert, cacertBytes, err = testutils.CertSetup()
	validCACert = string(cacertBytes)
	Expect(err).ToNot(HaveOccurred())

})
