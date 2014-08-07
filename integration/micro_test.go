package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

var _ = Describe("bosh-micro", func() {
	var (
		deploymentManifestDir      string
		deploymentManifestFilePath string
		cpiReleaseDir              string
		cpiReleaseFilename         string
	)

	Context("when a CPI release exists", func() {
		BeforeEach(func() {
			var err error
			cpiReleaseDir, err = ioutil.TempDir("", "integration-cpiRelease")
			Expect(err).NotTo(HaveOccurred())

			logger := boshlog.NewLogger(boshlog.LevelNone)
			fs := boshsys.NewOsFileSystem(logger)
			cpiReleaseFilename = path.Join(cpiReleaseDir, "cpi-release.tar")
			err = bmtestutils.GenerateCPIRelease(fs, cpiReleaseFilename)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := os.RemoveAll(cpiReleaseDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when a manifest exists", func() {
			BeforeEach(func() {
				var err error
				deploymentManifestDir, err = ioutil.TempDir("", "integration-deploymentManifest")
				Expect(err).NotTo(HaveOccurred())

				deploymentManifestFilePath = path.Join(deploymentManifestDir, "micro_deployment.yml")
				err = bmtestutils.GenerateDeploymentManifest(deploymentManifestFilePath)
				Expect(err).NotTo(HaveOccurred())
			})
			AfterEach(func() {
				err := os.RemoveAll(deploymentManifestDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("says the current deployment is set", func() {
				session, err := bmtestutils.RunBoshMicro("deployment", deploymentManifestFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(session.ExitCode()).To(Equal(0))

				session, err = bmtestutils.RunBoshMicro("deployment")
				Expect(err).NotTo(HaveOccurred())
				Expect(session.ExitCode()).To(Equal(0))
				Expect(session.Out.Contents()).To(ContainSubstring(
					fmt.Sprintf("Current deployment is '%s'", deploymentManifestFilePath)))
			})

			It("can deploy with a given CPI", func() {
				session, err := bmtestutils.RunBoshMicro("deployment", deploymentManifestFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(session.ExitCode()).To(Equal(0))

				session, err = bmtestutils.RunBoshMicro("deploy", cpiReleaseFilename)
				Expect(err).NotTo(HaveOccurred())
				Expect(session.ExitCode()).To(Equal(0))
			})
		})

		Context("when no manifest has been set", func() {
			It("refuses to deploy", func() {
				session, err := bmtestutils.RunBoshMicro("deploy", cpiReleaseFilename)
				Expect(err).NotTo(HaveOccurred())
				Expect(session.Err.Contents()).To(ContainSubstring("No deployment set"))
				Expect(session.ExitCode()).To(Equal(1))
			})
		})
	})

	Context("when no manifest has been set", func() {
		It("says deployment is not set", func() {
			session, err := bmtestutils.RunBoshMicro("deployment")
			Expect(err).NotTo(HaveOccurred())
			Expect(session.Err.Contents()).To(ContainSubstring("No deployment set"))
			Expect(session.ExitCode()).To(Equal(1))
		})
	})
})
