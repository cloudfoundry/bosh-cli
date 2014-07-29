package integration_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/integration/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bosh-micro deployment <manifest-filepath>", func() {
	var (
		deploymentManifestFilePath string
		session                    *gexec.Session
	)

	Context("valid manifest file", func() {
		BeforeEach(func() {
			deploymentManifestFilePath = "./test_fixtures/dummy.yml"
			session = RunBoshMicro("deployment", deploymentManifestFilePath)
			Expect(session.ExitCode()).Should(Equal(0))
		})

		XIt("is successfully accepts a valid manifest file", func() {
			Expect(session.Out.Contents()).To(ContainSubstring(deploymentManifestFilePath))
		})
	})
})
