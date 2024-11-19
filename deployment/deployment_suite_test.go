package deployment_test

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
