package integration_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeSuite(func() {
		err := bmtestutils.BuildExecutable()
		Expect(err).NotTo(HaveOccurred())
	})

	var (
		boshMicroPath string
		oldHome       string
	)
	BeforeEach(func() {
		oldHome = os.Getenv("HOME")

		var err error
		boshMicroPath, err = ioutil.TempDir("", "micro-bosh-cli-integration")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("HOME", boshMicroPath)
	})
	AfterEach(func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(boshMicroPath)
	})

	RunSpecs(t, "bosh-micro-cli Integration Suite")
}
