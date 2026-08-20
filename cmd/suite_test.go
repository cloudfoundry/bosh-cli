package cmd_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

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

// secureTLSClientMatcher is a Gomega matcher that asserts an *http.Client
// has TLS certificate verification enabled (InsecureSkipVerify == false).
// Use SecureTLSClientMatcher() to obtain an instance.
type secureTLSClientMatcher struct{}

func SecureTLSClientMatcher() types.GomegaMatcher {
	return secureTLSClientMatcher{}
}

func (m secureTLSClientMatcher) Match(actual interface{}) (bool, error) {
	client, ok := actual.(*http.Client)
	if !ok {
		return false, fmt.Errorf("SecureTLSClientMatcher expects a *http.Client, got %T", actual)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		return false, nil
	}
	return transport.TLSClientConfig != nil && !transport.TLSClientConfig.InsecureSkipVerify, nil
}

func (m secureTLSClientMatcher) FailureMessage(actual interface{}) string {
	return "Expected\n\t" + fmt.Sprintf("%#v", actual) + "\nto be a secure *http.Client with TLS certificate verification enabled (InsecureSkipVerify=false)"
}

func (m secureTLSClientMatcher) NegatedFailureMessage(actual interface{}) string {
	return "Expected\n\t" + fmt.Sprintf("%#v", actual) + "\nnot to be a secure *http.Client with TLS certificate verification enabled (InsecureSkipVerify=false)"
}
