package disk_test

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
	validCACert string
)
var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	cert, cacertBytes, err = testutils.Certsetup()
	validCACert = string(cacertBytes)
	Expect(err).ToNot(HaveOccurred())
	return []byte{}
}, func(in []byte) {})

func TestDisk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Disk Suite")
}
