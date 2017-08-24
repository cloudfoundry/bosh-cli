package grpcacl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGRPCACL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPC ACL Suite")
}
