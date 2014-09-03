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

type cpiRelease struct {
	releaseDir string
	fs         boshsys.FileSystem
}

type compilePackages struct {
	deploymentWorkspacePath string
	fs                      boshsys.FileSystem
}

func NewCompilePackages(deploymentWorkspacePath string, fs boshsys.FileSystem) compilePackages{
	return compilePackages{deploymentWorkspacePath: deploymentWorkspacePath, fs: fs}
}

func (c compilePackages) GetPackageBlobByName(packageName string) (blobReader, bool) {
	indexFile := path.Join(c.deploymentWorkspacePath, "index.json")
	Expect(c.fs.FileExists(indexFile)).To(BeTrue(), fmt.Sprintf("Expect index file to exist at %s", indexFile))

	index, err := c.fs.ReadFile(indexFile)
	Expect(err).NotTo(HaveOccurred())

	indexContent := IndexFile{}
	err = json.Unmarshal(index, &indexContent)
	Expect(err).NotTo(HaveOccurred())

 	blobId, found := c.getPackageBlobId(indexContent, packageName)
	if !found {
		return blobReader{}, false
	}

	return blobReader{path.Join(c.deploymentWorkspacePath, "blobs", blobId)}, true
}

func (c compilePackages) getPackageBlobId(indexContent IndexFile, packageName string) (string, bool) {
	for _, item := range indexContent {
		if item.Key.PackageName == packageName {
			return item.Value.BlobID, true
		}
	}

	return "", false
}

type blobReader struct {
	blobPath string
}

func (b blobReader) FileExists(fileName string) bool {
	session, err := bmtestutils.RunCommand("tar", "-tf", b.blobPath, fileName)
	Expect(err).ToNot(HaveOccurred())
	return session.ExitCode() == 0
}

func (b blobReader) FileContents(fileName string) []byte {
	session, err := bmtestutils.RunCommand("tar", "--to-stdout", "-xf", b.blobPath,  fileName)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(Equal(0))
	return session.Out.Contents()
}

func (c cpiRelease) createRelease() string {
	cmd := exec.Command("bosh", "create", "release", "--with-tarball")
	cmd.Dir = c.releaseDir

	session, err := bmtestutils.RunComplexCommand(cmd)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(Equal(0))

	re := regexp.MustCompile(`Release tarball.*: (.*)`)
	matches := re.FindStringSubmatch(string(session.Out.Contents()))
	return matches[1]
}

func (c cpiRelease) removeJob(jobName string) {
	c.fs.RemoveAll(path.Join(c.releaseDir, "jobs", jobName))
}

var _ = Describe("bosh-micro", func() {
	var (
		workspaceDir               string
		releaseTarball             string
		fs                         boshsys.FileSystem
		deploymentManifestFilePath string
		cpiRel                 cpiRelease
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
		cpiRel = cpiRelease{releaseDir, fs}
	})

	AfterEach(func() {
		err := os.RemoveAll(workspaceDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the CPI release is valid", func() {
		BeforeEach(func() {
			releaseTarball = cpiRel.createRelease()
		})

		It("compiles packages", func() {
			session, err := bmtestutils.RunBoshMicro("deploy", releaseTarball)
			Expect(err).ToNot(HaveOccurred())
			Expect(session.ExitCode()).To(Equal(0))

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring("Started compiling packages > dependency_package"))
			Expect(output).To(ContainSubstring("Started compiling packages > compiled_package"))
		})

		It("creates blobs with result of the compilation", func() {
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

			deploymentWorkspacePath := filepath.Join(os.Getenv("HOME"), ".bosh_micro", deploymentFile.UUID)
			compilePackages := NewCompilePackages(deploymentWorkspacePath, fs)
			blob, found := compilePackages.GetPackageBlobByName("compiled_package")
			Expect(found).To(BeTrue())
			Expect(blob.FileExists("compiled_file")).To(BeTrue())
		})
	})

	Context("when the CPI release is invalid", func() {
		var invalidCpiReleasePath string

		BeforeEach(func() {
			cpiRel.removeJob("cpi")
			invalidCpiReleasePath = cpiRel.createRelease()
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
