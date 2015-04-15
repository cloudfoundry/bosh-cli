package integration_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bitestutils "github.com/cloudfoundry/bosh-init/testutils"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	BeforeSuite(func() {
		err := bitestutils.BuildExecutable()
		Expect(err).NotTo(HaveOccurred())
	})

	var (
		homePath string
		oldHome  string
	)
	BeforeEach(func() {
		oldHome = os.Getenv("HOME")

		var err error
		homePath, err = ioutil.TempDir("", "bosh-init-cli-integration")
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("HOME", homePath)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Setenv("HOME", oldHome)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(homePath)
		Expect(err).NotTo(HaveOccurred())
	})

	RunSpecs(t, "Integration Suite")
}
