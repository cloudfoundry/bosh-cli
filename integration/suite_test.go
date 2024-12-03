package integration_test

import (
	"crypto/tls"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/testutils"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var (
	testHome string

	buildHTTPSServerCert        tls.Certificate
	buildHTTPSServerValidCACert string

	fs boshsys.FileSystem

	ui         *fakeui.FakeUI
	deps       cmd.BasicDeps
	cmdFactory cmd.Factory
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "integration")
}

var _ = BeforeSuite(func() {
	err := testutils.BuildExecutable()
	Expect(err).NotTo(HaveOccurred())

	var cacertBytes []byte
	buildHTTPSServerCert, cacertBytes, err = testutils.CertSetup()
	Expect(err).ToNot(HaveOccurred())

	buildHTTPSServerValidCACert = string(cacertBytes)
})

var _ = BeforeEach(func() {
	testHome = GinkgoT().TempDir()
	GinkgoT().Setenv("HOME", testHome)

	logger := boshlog.NewWriterLogger(boshlog.LevelNone, GinkgoWriter)
	fs = boshsys.NewOsFileSystem(logger)

	ui = &fakeui.FakeUI{}
	deps = cmd.NewBasicDepsWithFS(boshui.NewWrappingConfUI(ui, logger), fs, logger)

	cmdFactory = cmd.NewFactory(deps)
})

func buildHTTPSServer() (string, *ghttp.Server) {
	GinkgoHelper()

	server := ghttp.NewUnstartedServer()
	server.HTTPTestServer.TLS = &tls.Config{
		Certificates: []tls.Certificate{buildHTTPSServerCert},
	}

	server.HTTPTestServer.StartTLS()

	return buildHTTPSServerValidCACert, server
}

func createCommand(commandFactory cmd.Factory, args []string) cmd.Cmd {
	GinkgoHelper()
	command, err := commandFactory.New(args)
	Expect(err).ToNot(HaveOccurred())

	return command
}

func createAndExecCommand(commandFactory cmd.Factory, args []string) {
	GinkgoHelper()

	err := createCommand(commandFactory, args).Execute()
	Expect(err).ToNot(HaveOccurred())
}

func removeSHA1s(contents string) string {
	matchSHA1s := regexp.MustCompile("sha256:[a-z0-9]{64}")
	return matchSHA1s.ReplaceAllString(contents, "sha256:replaced")
}

func removeSHA256s(contents string) string {
	matchSHA256s := regexp.MustCompile("sha1: sha256:[a-z0-9]{64}\n")
	return matchSHA256s.ReplaceAllString(contents, "sha1: replaced\n")
}

func listTarballContents(tarballPath string) []string {
	contents := []string{}
	cmd := exec.Command("tar", "ztf", tarballPath)
	output, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	files := strings.Split(string(output), "\n")
	for _, file := range files {
		if file != "" {
			contents = append(contents, file)
		}
	}
	return contents
}

func setupReleaseDir(releaseDir, releaseName string) {
	By("running `init-release`, `generate-job`, and `generate-package`", func() {
		createAndExecCommand(cmdFactory, []string{"init-release", "--git", "--dir", releaseDir})
		createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", releaseDir})
		createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", releaseDir})
	})

	By("creating a job that depends on `pkg1`", func() {
		jobSpecPath := filepath.Join(releaseDir, "jobs", "job1", "spec")

		contents, err := fs.ReadFileString(jobSpecPath)
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
		Expect(err).ToNot(HaveOccurred())
	})

	By("adding some content", func() {
		err := fs.WriteFileString(filepath.Join(releaseDir, "src", "in-src"), "in-src")
		Expect(err).ToNot(HaveOccurred())

		pkg1SpecPath := filepath.Join(releaseDir, "packages", "pkg1", "spec")

		contents, err := fs.ReadFileString(pkg1SpecPath)
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "files: []", "files:\n- in-src", -1))
		Expect(err).ToNot(HaveOccurred())
	})

	By("creating a release with local blobstore", func() {
		blobstoreDir := filepath.Join(releaseDir, ".blobstore")

		err := fs.MkdirAll(blobstoreDir, 0777)
		Expect(err).ToNot(HaveOccurred())

		finalYaml := "name: " + releaseName + `
blobstore:
  provider: local
  options:
    blobstore_path: ` + blobstoreDir

		err = fs.WriteFileString(filepath.Join(releaseDir, "config", "final.yml"), finalYaml)
		Expect(err).ToNot(HaveOccurred())
	})
}

func createSimpleRelease() string {
	tmpDir, err := fs.TempDir("bosh-create-release-int-test")
	Expect(err).ToNot(HaveOccurred())

	relName := filepath.Base(tmpDir)

	By("running `create-release`", func() {
		createAndExecCommand(cmdFactory, []string{"init-release", "--dir", tmpDir})
		Expect(fs.FileExists(filepath.Join(tmpDir, "config"))).To(BeTrue())
		Expect(fs.FileExists(filepath.Join(tmpDir, "jobs"))).To(BeTrue())
		Expect(fs.FileExists(filepath.Join(tmpDir, "packages"))).To(BeTrue())
		Expect(fs.FileExists(filepath.Join(tmpDir, "src"))).To(BeTrue())
	})

	By("running `generate-job`", func() {
		createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", tmpDir})
	})

	By("running `generate-package`", func() {
		createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", tmpDir})
	})

	err = fs.WriteFileString(filepath.Join(tmpDir, "LICENSE"), "LICENSE")
	Expect(err).ToNot(HaveOccurred())

	By("by adding a package spec file", func() {
		pkg1SpecPath := filepath.Join(tmpDir, "packages", "pkg1", "spec")

		contents, err := fs.ReadFileString(pkg1SpecPath)
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "dependencies: []", "dependencies: []", -1))
		Expect(err).ToNot(HaveOccurred())
	})

	By("by adding a job spec file", func() {
		jobSpecPath := filepath.Join(tmpDir, "jobs", "job1", "spec")

		contents, err := fs.ReadFileString(jobSpecPath)
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
		Expect(err).ToNot(HaveOccurred())
	})

	sha2ifyReleasePath := filepath.Join(tmpDir, "sha2ify-release.tgz")

	By("running `create-release`", func() { // Make empty release
		createAndExecCommand(cmdFactory, []string{"create-release", "--sha2", "--dir", tmpDir, "--tarball", sha2ifyReleasePath})

		Expect(fs.FileExists(filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.1.yml"))).To(BeTrue())
	})

	return sha2ifyReleasePath
}
