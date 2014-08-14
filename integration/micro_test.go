package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

var _ = Describe("bosh-micro", func() {
	var (
		deploymentManifestDir      string
		deploymentManifestFilePath string
		cpiReleasePath             string
		fileSystem                 boshsys.FileSystem
	)

	Context("when a CPI release exists", func() {
		BeforeEach(func() {
			var err error
			cpiReleasePath = testCpiFilePath
			Expect(err).NotTo(HaveOccurred())
			logger := boshlog.NewLogger(boshlog.LevelNone)
			fileSystem = boshsys.NewOsFileSystem(logger)
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
					fmt.Sprintf("Current deployment is `%s'", deploymentManifestFilePath)))
			})

			It("can deploy with a given CPI", func() {
				session, err := bmtestutils.RunBoshMicro("deployment", deploymentManifestFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(session.ExitCode()).To(Equal(0))

				session, err = bmtestutils.RunBoshMicro("deploy", cpiReleasePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(session.ExitCode()).To(Equal(0))
				boshMicroHiddenPath := filepath.Join(os.Getenv("HOME"), ".bosh_micro")
				Expect(fileSystem.FileExists(boshMicroHiddenPath)).To(BeTrue())
			})

			Context("when the CPI release is invalid", func() {
				var invalidCpiReleasePath string

				BeforeEach(func() {
					var err error
					invalidCpiReleasePath, err = bmtestutils.DownloadTestCpiRelease(
						"https://s3.amazonaws.com/bosh-dependencies/invalid_cpi_release.tgz")
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.Remove(invalidCpiReleasePath)
				})

				It("says CPI release is invalid", func() {
					session, err := bmtestutils.RunBoshMicro("deployment", deploymentManifestFilePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(session.ExitCode()).To(Equal(0))

					Expect(err).NotTo(HaveOccurred())

					session, err = bmtestutils.RunBoshMicro("deploy", invalidCpiReleasePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(session.Err.Contents()).To(ContainSubstring("is not a valid CPI release"))
					Expect(session.ExitCode()).To(Equal(1))
				})
			})
		})

		Context("when no manifest has been set", func() {
			It("refuses to deploy", func() {
				session, err := bmtestutils.RunBoshMicro("deploy", cpiReleasePath)
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
