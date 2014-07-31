package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/integration"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeSuite(BuildExecutable)

	StubBoshMicroPath()

	RunSpecs(t, "bosh-micro-cli Integration Suite")
}
