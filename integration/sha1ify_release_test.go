package integration_test

import (
	"path/filepath"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("sha1ify-release", func() {

	var (
		ui                  *fakeui.FakeUI
		fs                  boshsys.FileSystem
		deps                cmd.BasicDeps
		cmdFactory          cmd.Factory
		releaseProvider     boshrel.Provider
		createSimpleRelease func() string
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		confUI := boshui.NewWrappingConfUI(ui, logger)

		fs = boshsys.NewOsFileSystem(logger)
		deps = cmd.NewBasicDepsWithFS(confUI, fs, logger)
		cmdFactory = cmd.NewFactory(deps)

		releaseProvider = boshrel.NewProvider(
			deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)

	})

	It("converts the SHA2s into SHA1s for packages and jobs", func() {
		sha1ifyReleasePath := createSimpleRelease()
		defer fs.RemoveAll(filepath.Dir(sha1ifyReleasePath)) //nolint:errcheck

		dirtyPath, err := fs.TempDir("sha1release")
		Expect(err).ToNot(HaveOccurred())

		outFile := filepath.Join(dirtyPath, "small-sha1-release.tgz")

		createAndExecCommand(cmdFactory, []string{"sha1ify-release", sha1ifyReleasePath, outFile})

		extractor := releaseProvider.NewExtractingArchiveReader()

		release, err := extractor.Read(outFile)
		Expect(err).ToNot(HaveOccurred())

		By("keeping all the jobs and packages")
		Expect(release.Jobs()).To(HaveLen(1))
		Expect(release.Packages()).To(HaveLen(1))
		Expect(release.License()).ToNot(BeNil())

		By("converting the SHAs to SHA-1")
		jobArchiveSha := release.Jobs()[0].ArchiveDigest()
		Expect(jobArchiveSha).To(HaveLen(40))

		packageArchiveSha := release.Packages()[0].ArchiveDigest()
		Expect(packageArchiveSha).To(HaveLen(40))

		licenseArchiveSha := release.License().ArchiveDigest()
		Expect(licenseArchiveSha).To(HaveLen(40))

		By("preserving the version string exactly")
		Expect(release.Version()).To(Equal("0+dev.1"))
	})

	It("converts the SHA2s into SHA1s for packages and jobs", func() {
		dirtyPath, err := fs.TempDir("sha2release")
		Expect(err).ToNot(HaveOccurred())

		outFile := filepath.Join(dirtyPath, "small-sha1-release.tgz")

		createAndExecCommand(cmdFactory, []string{"sha1ify-release", "assets/small-sha256-compiled-release.tgz", outFile})

		extractor := releaseProvider.NewExtractingArchiveReader()

		release, err := extractor.Read(outFile)
		Expect(err).ToNot(HaveOccurred())

		By("keeping all the jobs and packages")
		Expect(release.Jobs()).To(HaveLen(1))
		Expect(release.CompiledPackages()).To(HaveLen(1))

		By("converting the SHAs to SHA-1")
		jobArchiveSha := release.Jobs()[0].ArchiveDigest()
		Expect(jobArchiveSha).To(HaveLen(40))
		compiledPackageSha := release.CompiledPackages()[0].ArchiveDigest()
		Expect(compiledPackageSha).To(HaveLen(40))

		By("preserving the version string exactly")
		Expect(release.Version()).To(Equal("0+dev.3"))
	})

	createSimpleRelease = func() string {
		tmpDir, err := fs.TempDir("bosh-create-release-int-test")
		Expect(err).ToNot(HaveOccurred())

		relName := filepath.Base(tmpDir)

		{
			createAndExecCommand(cmdFactory, []string{"init-release", "--dir", tmpDir})
			Expect(fs.FileExists(filepath.Join(tmpDir, "config"))).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(tmpDir, "jobs"))).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(tmpDir, "packages"))).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(tmpDir, "src"))).To(BeTrue())
		}

		createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", tmpDir})
		createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", tmpDir})

		err = fs.WriteFileString(filepath.Join(tmpDir, "LICENSE"), "LICENSE")
		Expect(err).ToNot(HaveOccurred())

		{
			pkg1SpecPath := filepath.Join(tmpDir, "packages", "pkg1", "spec")

			contents, err := fs.ReadFileString(pkg1SpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "dependencies: []", "dependencies: []", -1))
			Expect(err).ToNot(HaveOccurred())
		}

		{
			jobSpecPath := filepath.Join(tmpDir, "jobs", "job1", "spec")

			contents, err := fs.ReadFileString(jobSpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
			Expect(err).ToNot(HaveOccurred())
		}

		sha2ifyReleasePath := filepath.Join(tmpDir, "sha2ify-release.tgz")

		{ // Make empty release
			createAndExecCommand(cmdFactory, []string{"create-release", "--sha2", "--dir", tmpDir, "--tarball", sha2ifyReleasePath})

			_, err := fs.ReadFileString(filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.1.yml"))
			Expect(err).ToNot(HaveOccurred())
		}

		return sha2ifyReleasePath
	}
})
