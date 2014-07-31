package integration_test

import (
	"github.com/cloudfoundry/bosh-micro-cli/integration/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeSuite(test_helpers.BuildExecutable)

	test_helpers.StubBoshMicroPath()

	RunSpecs(t, "bosh-micro-cli Integration Suite")
}
