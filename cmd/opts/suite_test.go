package opts_test

import (
	"crypto/tls"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/testutils"
)

var (
	cert        tls.Certificate
	cacertBytes []byte
	validCACert string
)
var _ = BeforeSuite(func() {
	var err error
	cert, cacertBytes, err = testutils.CertSetup()
	validCACert = string(cacertBytes)
	Expect(err).ToNot(HaveOccurred())
})

func TestReg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cmd/opts")
}
