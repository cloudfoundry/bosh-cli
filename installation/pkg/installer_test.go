package pkg_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebminstallblob "github.com/cloudfoundry/bosh-init/installation/blob/fakes"

	. "github.com/cloudfoundry/bosh-init/installation/pkg"
)

var _ = Describe("PackageInstaller", func() {
	var (
		installer     Installer
		blobExtractor *fakebminstallblob.FakeExtractor
		targetDir     string
	)

	BeforeEach(func() {
		blobExtractor = fakebminstallblob.NewFakeExtractor()
		targetDir = "fake-target-dir"
		installer = NewPackageInstaller(blobExtractor)
	})

	Describe("Install", func() {
		var (
			compiledPackageRef CompiledPackageRef
		)

		BeforeEach(func() {
			compiledPackageRef = CompiledPackageRef{
				Name:        "fake-package-name",
				Version:     "fake-package-version", // unused
				BlobstoreID: "fake-blob-id",
				SHA1:        "fake-package-fingerprint",
			}
			blobExtractor.SetExtractBehavior("fake-blob-id", "fake-package-fingerprint", "fake-target-dir/fake-package-name", nil)
		})

		It("extracts the blob into the target dir", func() {
			err := installer.Install(compiledPackageRef, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobExtractor.ExtractInputs).To(ContainElement(fakebminstallblob.ExtractInput{
				BlobID:    "fake-blob-id",
				BlobSHA1:  "fake-package-fingerprint",
				TargetDir: "fake-target-dir/fake-package-name",
			}))
		})

		Context("when extracting errors", func() {
			BeforeEach(func() {
				blobExtractor.SetExtractBehavior("fake-blob-id", "fake-package-fingerprint", "fake-target-dir/fake-package-name", errors.New("fake-extract-error"))
			})

			It("returns an error", func() {
				err := installer.Install(compiledPackageRef, targetDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-extract-error"))
			})
		})
	})
})
