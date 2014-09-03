package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type Key struct {
	PackageName        string
	PackageFingerprint string
}

type Value struct {
	BlobID   string
	BlobSha1 string
}

type Item struct {
	Key   Key
	Value Value
}

type IndexFile []Item

type DeploymentFile struct {
	UUID string
}

type CpiRelease struct {
	releaseDir string
	fs         boshsys.FileSystem
}

func (c CpiRelease) createRelease() string {
	cmd := exec.Command("bosh", "create", "release", "--with-tarball")
	cmd.Dir = c.releaseDir

	session, err := bmtestutils.RunComplexCommand(cmd)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(Equal(0))

	re := regexp.MustCompile(`Release tarball.*: (.*)`)
	matches := re.FindStringSubmatch(string(session.Out.Contents()))
	return matches[1]
}

func (c CpiRelease) removeJob(jobName string) {
	c.fs.RemoveAll(path.Join(c.releaseDir, "jobs", jobName))
}

var _ = Describe("bosh-micro", func() {
	var (
		workspaceDir               string
		releaseTarball             string
		fs                         boshsys.FileSystem
		deploymentManifestFilePath string
		cpiRelease                 CpiRelease
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = boshsys.NewOsFileSystem(logger)

		var err error
		workspaceDir, err = fs.TempDir("bosh-micro-intergration")
		Expect(err).NotTo(HaveOccurred())

		deploymentManifestFilePath = path.Join(workspaceDir, "micro_deployment.yml")
		err = bmtestutils.GenerateDeploymentManifest(deploymentManifestFilePath)
		Expect(err).NotTo(HaveOccurred())

		session, err := bmtestutils.RunBoshMicro("deployment", deploymentManifestFilePath)
		Expect(err).NotTo(HaveOccurred())
		Expect(session.ExitCode()).To(Equal(0))

		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())

		assetDir := filepath.Join(pwd, "../Fixtures", "test_release")
		session, err = bmtestutils.RunCommand("cp", "-r", assetDir, workspaceDir)
		Expect(err).ToNot(HaveOccurred())
		Expect(session.ExitCode()).To(Equal(0))

		releaseDir := filepath.Join(workspaceDir, "test_release")
		cpiRelease = CpiRelease{releaseDir, fs}
	})

	AfterEach(func() {
		err := os.RemoveAll(workspaceDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the CPI release is valid", func() {
		BeforeEach(func() {
			releaseTarball = cpiRelease.createRelease()
		})

		It("compiles packages", func() {
			session, err := bmtestutils.RunBoshMicro("deploy", releaseTarball)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode()).To(Equal(0))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("Started compiling packages > dependency_package"))
			Expect(output).To(ContainSubstring("Started compiling packages > compiled_package"))
		})

		It("creates blobs", func() {
			session, err := bmtestutils.RunBoshMicro("deploy", releaseTarball)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode()).To(Equal(0))

			deploymentFilePath := path.Join(workspaceDir, "deployment.json")
			Expect(fs.FileExists(deploymentFilePath)).To(BeTrue())

			deploymentRawContent, err := fs.ReadFile(deploymentFilePath)
			Expect(err).NotTo(HaveOccurred())

			deploymentFile := DeploymentFile{}
			err = json.Unmarshal(deploymentRawContent, &deploymentFile)
			Expect(err).NotTo(HaveOccurred())

			boshMicroHiddenPath := filepath.Join(os.Getenv("HOME"), ".bosh_micro", deploymentFile.UUID)
			Expect(fs.FileExists(boshMicroHiddenPath)).To(BeTrue())
			indexFile := path.Join(boshMicroHiddenPath, "index.json")
			Expect(fs.FileExists(indexFile)).To(BeTrue(), fmt.Sprintf("Expect index file to exist at %s", indexFile))

			index, err := fs.ReadFile(path.Join(boshMicroHiddenPath, "index.json"))
			Expect(err).NotTo(HaveOccurred())

			indexContent := IndexFile{}
			err = json.Unmarshal(index, &indexContent)

			Expect(err).NotTo(HaveOccurred())
			for _, item := range indexContent {
				Expect(item.Value.BlobSha1).ToNot(BeEmpty())
			}
		})
	})

	Context("when the CPI release is invalid", func() {
		var invalidCpiReleasePath string

		BeforeEach(func() {
			cpiRelease.removeJob("cpi")
			invalidCpiReleasePath = cpiRelease.createRelease()
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
