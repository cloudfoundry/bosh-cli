package integration_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/integration"
)

var _ = Describe("bosh-micro", func() {
	var (
		deploymentManifestFilePath string
	)

	Context("when a manifest exists", func() {
		BeforeEach(func() {
			deploymentManifestFilePath = GenerateDeploymentManifest()
		})
		AfterEach(func() {
			err := os.RemoveAll(deploymentManifestFilePath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("says the current deployment is set", func() {
			session := RunBoshMicro("deployment", deploymentManifestFilePath)
			Expect(session.ExitCode()).To(Equal(0))

			session = RunBoshMicro("deployment")
			Expect(session.ExitCode()).To(Equal(0))
			Expect(session.Out.Contents()).To(ContainSubstring(
				fmt.Sprintf("Current deployment is '%s'", deploymentManifestFilePath)))
		})
	})

	Context("when no manifest has been set", func() {
		It("says deployment is not set", func() {
			session := RunBoshMicro("deployment")
			Expect(session.Err.Contents()).To(ContainSubstring("No deployment set"))
			Expect(session.ExitCode()).To(Equal(1))
		})
	})
})
