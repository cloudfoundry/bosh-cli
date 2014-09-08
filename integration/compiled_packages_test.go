package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type PackageKey struct {
	PackageName        string
	PackageFingerprint string
}

type PackageValue struct {
	BlobID   string
	BlobSha1 string
}

type PackageItem struct {
	Key   PackageKey
	Value PackageValue
}

type CompiledPackagesIndexFile []PackageItem

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

func NewCompilePackages(deploymentWorkspacePath string, fs boshsys.FileSystem) compilePackages {
	return compilePackages{deploymentWorkspacePath: deploymentWorkspacePath, fs: fs}
}

func (c compilePackages) GetPackageBlobByName(packageName string) (blobReader, bool) {
	indexFile := path.Join(c.deploymentWorkspacePath, "compiled_packages.json")
	Expect(c.fs.FileExists(indexFile)).To(BeTrue(), fmt.Sprintf("Expect index file to exist at %s", indexFile))

	index, err := c.fs.ReadFile(indexFile)
	Expect(err).NotTo(HaveOccurred())

	indexContent := CompiledPackagesIndexFile{}
	err = json.Unmarshal(index, &indexContent)
	Expect(err).NotTo(HaveOccurred())

	blobID, found := c.getPackageBlobID(indexContent, packageName)
	if !found {
		return blobReader{}, false
	}

	return blobReader{path.Join(c.deploymentWorkspacePath, "blobs", blobID)}, true
}

func (c compilePackages) getPackageBlobID(indexContent CompiledPackagesIndexFile, packageName string) (string, bool) {
	for _, item := range indexContent {
		if item.Key.PackageName == packageName {
			return item.Value.BlobID, true
		}
	}

	return "", false
}

type RenderedTemplateKey struct {
	JobName        string
	JobFingerprint string
}

type RenderedTemplateValue struct {
	BlobID   string
	BlobSha1 string
}

type RenderedTemplateItem struct {
	Key   RenderedTemplateKey
	Value RenderedTemplateValue
}

type RenderedTemplatesIndexFile []RenderedTemplateItem

type renderedTemplates struct {
	deploymentWorkspacePath string
	fs                      boshsys.FileSystem
}

func NewRenderedTemplates(deploymentWorkspacePath string, fs boshsys.FileSystem) renderedTemplates {
	return renderedTemplates{deploymentWorkspacePath: deploymentWorkspacePath, fs: fs}
}

func (c renderedTemplates) GetRenderedTemplateBlobByName(templateName string) (blobReader, bool) {
	indexFile := path.Join(c.deploymentWorkspacePath, "templates.json")
	Expect(c.fs.FileExists(indexFile)).To(BeTrue(), fmt.Sprintf("Expect index file to exist at %s", indexFile))

	index, err := c.fs.ReadFile(indexFile)
	Expect(err).NotTo(HaveOccurred())

	indexContent := RenderedTemplatesIndexFile{}
	err = json.Unmarshal(index, &indexContent)
	Expect(err).NotTo(HaveOccurred())

	blobID, found := c.getTemplateBlobID(indexContent, templateName)
	if !found {
		return blobReader{}, false
	}

	return blobReader{path.Join(c.deploymentWorkspacePath, "blobs", blobID)}, true
}

func (c renderedTemplates) getTemplateBlobID(indexContent RenderedTemplatesIndexFile, jobName string) (string, bool) {
	for _, item := range indexContent {
		if item.Key.JobName == jobName {
			return item.Value.BlobID, true
		}
	}

	return "", false
}

type deploymentWorkspace struct {
	workspaceDir string
	fs           boshsys.FileSystem
}

func (d deploymentWorkspace) Path() string {
	deploymentFilePath := path.Join(d.workspaceDir, "deployment.json")
	Expect(d.fs.FileExists(deploymentFilePath)).To(BeTrue())

	deploymentRawContent, err := d.fs.ReadFile(deploymentFilePath)
	Expect(err).NotTo(HaveOccurred())

	deploymentFile := DeploymentFile{}
	err = json.Unmarshal(deploymentRawContent, &deploymentFile)
	Expect(err).NotTo(HaveOccurred())

	return filepath.Join(os.Getenv("HOME"), ".bosh_micro", deploymentFile.UUID)
}

type blobReader struct {
	blobPath string
}

func (b blobReader) FileExists(fileName string) bool {
	session, err := bmtestutils.RunCommand("tar", "-tf", b.blobPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(session.ExitCode()).To(Equal(0))
	return strings.Contains(string(session.Out.Contents()), fileName)
}

func (b blobReader) FileContents(fileName string) []byte {
	session, err := bmtestutils.RunCommand("tar", "--to-stdout", "-xf", b.blobPath, path.Join(".", fileName))
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
		cpiRel                     cpiRelease
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = boshsys.NewOsFileSystem(logger)

		var err error
		workspaceDir, err = fs.TempDir("bosh-micro-intergration")
		Expect(err).NotTo(HaveOccurred())

		deploymentManifestFilePath = path.Join(workspaceDir, "micro_deployment.yml")

		manifestContents := `
---
name: fake-deployment
cloud_provider:
  properties:
    fake_cpi_specified_property:
      second_level: fake_specified_property_value
`

		err = bmtestutils.GenerateDeploymentManifest(deploymentManifestFilePath, fs, manifestContents)
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

			workspace := deploymentWorkspace{workspaceDir, fs}
			compilePackages := NewCompilePackages(workspace.Path(), fs)
			blob, found := compilePackages.GetPackageBlobByName("compiled_package")
			Expect(found).To(BeTrue())
			Expect(blob.FileExists("compiled_file")).To(BeTrue())
		})

		It("renders job templates including network config", func() {
			session, err := bmtestutils.RunBoshMicro("deploy", releaseTarball)
			Expect(err).NotTo(HaveOccurred())
			Expect(session.ExitCode()).To(Equal(0))

			workspace := deploymentWorkspace{workspaceDir, fs}
			renderedTemplates := NewRenderedTemplates(workspace.Path(), fs)
			blob, found := renderedTemplates.GetRenderedTemplateBlobByName("cpi")
			Expect(found).To(BeTrue())
			Expect(blob.FileExists("bin/cpi")).To(BeTrue())
			Expect(blob.FileContents("bin/cpi")).To(ContainSubstring("fake_cpi_default_cmd fake_cpi_default_value"))
			Expect(blob.FileContents("bin/cpi")).To(ContainSubstring("fake_cpi_specified_cmd fake_specified_property_value"))
			Expect(blob.FileContents("bin/cpi")).To(ContainSubstring(`ip: ""`))
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
