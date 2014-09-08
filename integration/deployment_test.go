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
		fileSystem                 boshsys.FileSystem
	)

	Context("when a CPI release exists", func() {
		BeforeEach(func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			fileSystem = boshsys.NewOsFileSystem(logger)
		})

		Context("when a manifest exists", func() {
			BeforeEach(func() {
				var err error
				deploymentManifestDir, err = ioutil.TempDir("", "integration-deploymentManifest")
				Expect(err).NotTo(HaveOccurred())

				deploymentManifestFilePath = path.Join(deploymentManifestDir, "micro_deployment.yml")

				err = bmtestutils.GenerateDeploymentManifest(deploymentManifestFilePath, fileSystem, "---")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.RemoveAll(deploymentManifestDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not set up the workspace if there has not been a deployment", func() {
				session, err := bmtestutils.RunBoshMicro("deploy")
				Expect(session.ExitCode()).ToNot(Equal(0))

				boshMicroHiddenPath := filepath.Join(os.Getenv("HOME"), ".bosh_micro")
				filesInBoshMicro, err := fileSystem.Glob(path.Join(boshMicroHiddenPath, "*"))
				Expect(err).ToNot(HaveOccurred())

				Expect(len(filesInBoshMicro)).To(Equal(0))
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
		})

		Context("when no manifest has been set", func() {
			It("refuses to deploy", func() {
				tmpFile, err := fileSystem.TempFile("")
				Expect(err).NotTo(HaveOccurred())
				defer fileSystem.RemoveAll(tmpFile.Name())

				session, err := bmtestutils.RunBoshMicro("deploy", tmpFile.Name())
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
