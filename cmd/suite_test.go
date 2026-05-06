package cmd_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
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
	RunSpecs(t, "cmd")
}

// secureTLSClientMatcher is a gomock.Matcher that asserts an *http.Client
// has TLS certificate verification enabled (InsecureSkipVerify == false).
// Use SecureTLSClientMatcher() to obtain an instance.
type secureTLSClientMatcher struct{}

func SecureTLSClientMatcher() gomock.Matcher {
	return secureTLSClientMatcher{}
}

func (m secureTLSClientMatcher) Matches(x interface{}) bool {
	client, ok := x.(*http.Client)
	if !ok {
		return false
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		return false
	}
	return transport.TLSClientConfig != nil && !transport.TLSClientConfig.InsecureSkipVerify
}

func (m secureTLSClientMatcher) String() string {
	return "is a secure *http.Client with TLS certificate verification enabled (InsecureSkipVerify=false)"
}
