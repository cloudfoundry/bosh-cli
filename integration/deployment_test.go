package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/cloudfoundry/bosh-micro-cli/integration/test_helpers"
)

var _ = Describe("bosh-micro deployment <manifest-filepath>", func() {
	var (
		deploymentManifestFilePath string
		session                    *gexec.Session
	)

	Context("with a manifest file", func() {
		BeforeEach(func() {
			tmpFile, err := ioutil.TempFile("", "bosh-micro-cli")
			Expect(err).NotTo(HaveOccurred())
			deploymentManifestFilePath = tmpFile.Name()
		})

		Context("valid manifest file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				session := RunBoshMicro("deployment", deploymentManifestFilePath)
				Expect(session.ExitCode()).Should(Equal(0))
			})

			AfterEach(func() {
				err := os.RemoveAll(deploymentManifestFilePath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("is successfully accepts a valid manifest file", func() {
				session = RunBoshMicro("deployment")
				Expect(session.ExitCode()).Should(Equal(0))
				Expect(session.Out.Contents()).To(ContainSubstring(
					fmt.Sprintf("Current deployment is '%s'", deploymentManifestFilePath)))
			})
		})
	})

	Context("without a manifest file", func() {
		It("says deployment not set", func() {
			session := RunBoshMicro("deployment")
			Expect(session.Err.Contents()).To(ContainSubstring("Deployment not set"))
			Expect(session.ExitCode()).Should(Equal(1))
		})
	})
})
