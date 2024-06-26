package integration_test

import (
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
)

var _ = Describe("sha2ify-release", func() {

	var (
		releaseProvider     boshrel.Provider
		createSimpleRelease func() string
		removeSHA1s         func(string) string
	)

	BeforeEach(func() {
		releaseProvider =
			boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)

	})

	It("converts the SHA1s into SHA2s for packages and jobs", func() {
		sha2ifyReleasePath := createSimpleRelease()
		defer fs.RemoveAll(filepath.Dir(sha2ifyReleasePath)) //nolint:errcheck

		dirtyPath, err := fs.TempDir("sha2release")
		Expect(err).ToNot(HaveOccurred())

		outFile := filepath.Join(dirtyPath, "small-sha256-release.tgz")

		createAndExecCommand(cmdFactory, []string{"sha2ify-release", sha2ifyReleasePath, outFile})

		extractor := releaseProvider.NewExtractingArchiveReader()

		release, err := extractor.Read(outFile)
		Expect(err).ToNot(HaveOccurred())

		By("keeping all the jobs and packages")
		Expect(release.Jobs()).To(HaveLen(1))
		Expect(release.Packages()).To(HaveLen(1))
		Expect(release.License()).ToNot(BeNil())

		By("converting the SHAs to 256")
		jobArchiveSha := release.Jobs()[0].ArchiveDigest()
		Expect(removeSHA1s(jobArchiveSha)).To(Equal("sha256:replaced"))

		packageArchiveSha := release.Packages()[0].ArchiveDigest()
		Expect(removeSHA1s(packageArchiveSha)).To(Equal("sha256:replaced"))

		licenseArchiveSha := release.License().ArchiveDigest()
		Expect(removeSHA1s(licenseArchiveSha)).To(Equal("sha256:replaced"))

		By("preserving the version string exactly")
		Expect(release.Version()).To(Equal("0+dev.1"))
	})

	It("converts the SHA1s into SHA2s for packages and jobs", func() {
		dirtyPath, err := fs.TempDir("sha2release")
		Expect(err).ToNot(HaveOccurred())

		outFile := filepath.Join(dirtyPath, "small-sha256-release.tgz")

		createAndExecCommand(cmdFactory, []string{"sha2ify-release", "assets/small-sha128-compiled-release.tgz", outFile})

		extractor := releaseProvider.NewExtractingArchiveReader()

		release, err := extractor.Read(outFile)
		Expect(err).ToNot(HaveOccurred())

		By("keeping all the jobs and packages")
		Expect(release.Jobs()).To(HaveLen(1))
		Expect(release.CompiledPackages()).To(HaveLen(1))

		By("converting the SHAs to 256")
		jobArchiveSha := release.Jobs()[0].ArchiveDigest()
		Expect(removeSHA1s(jobArchiveSha)).To(Equal("sha256:replaced"))
		compiledPackageSha := release.CompiledPackages()[0].ArchiveDigest()
		Expect(removeSHA1s(compiledPackageSha)).To(Equal("sha256:replaced"))

		By("preserving the version string exactly")
		Expect(release.Version()).To(Equal("0+dev.3"))
	})

	removeSHA1s = func(contents string) string {
		matchSHA1s := regexp.MustCompile("sha256:[a-z0-9]{64}")
		return matchSHA1s.ReplaceAllString(contents, "sha256:replaced")
	}

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
			createAndExecCommand(cmdFactory, []string{"create-release", "--dir", tmpDir, "--tarball", sha2ifyReleasePath})

			_, err := fs.ReadFileString(filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.1.yml"))
			Expect(err).ToNot(HaveOccurred())
		}

		return sha2ifyReleasePath
	}
})
