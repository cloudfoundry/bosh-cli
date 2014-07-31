package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/cloudfoundry/bosh-micro-cli/integration"
)

var _ = Describe("bosh-micro", func() {
	var (
		deploymentManifestFilePath string
		session                    *gexec.Session
	)

	Context("when a manifest has been set", func() {
		BeforeEach(func() {
			tmpFile, err := ioutil.TempFile("", "bosh-micro-cli")
			Expect(err).NotTo(HaveOccurred())
			deploymentManifestFilePath = tmpFile.Name()

			err = ioutil.WriteFile(deploymentManifestFilePath, []byte(""), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := os.RemoveAll(deploymentManifestFilePath)
			Expect(err).NotTo(HaveOccurred())
		})

		BeforeEach(func() {
			session := RunBoshMicro("deployment", deploymentManifestFilePath)
			Expect(session.ExitCode()).Should(Equal(0))
		})

		It("says the current deployment is set", func() {
			session = RunBoshMicro("deployment")
			Expect(session.ExitCode()).Should(Equal(0))
			Expect(session.Out.Contents()).To(ContainSubstring(
				fmt.Sprintf("Current deployment is '%s'", deploymentManifestFilePath)))
		})
	})

	Context("when no manifest has been set", func() {
		It("says deployment is not set", func() {
			session := RunBoshMicro("deployment")
			Expect(session.Err.Contents()).To(ContainSubstring("Deployment not set"))
			Expect(session.ExitCode()).Should(Equal(1))
		})
	})
})
