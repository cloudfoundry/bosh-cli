package integration_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
)

var _ = Describe("sha2ify-release", func() {
	var releaseProvider boshrel.Provider

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
})
