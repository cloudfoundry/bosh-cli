package grpcrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGRPCRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPC Runner Suite")
}
