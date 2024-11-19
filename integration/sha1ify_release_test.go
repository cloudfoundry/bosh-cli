package integration_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
)

var _ = Describe("sha1ify-release", func() {
	var releaseProvider boshrel.Provider

	BeforeEach(func() {
		releaseProvider =
			boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
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
})
