package integration_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/integration/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bosh-micro deployment <manifest-filepath>", func() {
	var (
		deploymentManifestFilePath string
	)

	Context("valid manifest file", func() {
		BeforeEach(func() {
			session := RunBoshMicro("deployment", deploymentManifestFilePath)
			Ω(session.ExitCode()).Should(Equal(0))
		})

		It("is successfully accepts a valid manifest file", func() {

		})
	})
})
