package integration_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/integration/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"io/ioutil"
	"os"
)

var _ = Describe("bosh-micro deployment <manifest-filepath>", func() {
	var (
		deploymentManifestFilePath string
		session                    *gexec.Session
	)

	Context("valid manifest file", func() {
		BeforeEach(func() {
			tmpFile, err := ioutil.TempFile("", "bosh-micro-cli")
			Expect(err).NotTo(HaveOccurred())
			deploymentManifestFilePath = tmpFile.Name()

			err = ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			session = RunBoshMicro("deployment", deploymentManifestFilePath)
			Expect(session.ExitCode()).Should(Equal(0))
		})

		AfterEach(func() {
			err := os.RemoveAll(deploymentManifestFilePath)
			Expect(err).NotTo(HaveOccurred())
		})

		XIt("is successfully accepts a valid manifest file", func() {
			Expect(session.Out.Contents()).To(ContainSubstring(deploymentManifestFilePath))
		})
	})
})
