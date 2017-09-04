package requestid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRequestID(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Request ID Suite")
}
