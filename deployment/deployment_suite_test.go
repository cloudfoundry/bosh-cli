package deployment_test

import (
	"crypto/tls"
	"testing"

	"github.com/cloudfoundry/bosh-cli/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	cert        tls.Certificate
	cacertBytes []byte
)
var _ = BeforeSuite(func() {
	var err error
	cert, cacertBytes, err = testutils.CertSetup()
	Expect(err).ToNot(HaveOccurred())
})

func TestDeployment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deployment Suite")
}
