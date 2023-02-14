package license_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/release/license"
	boshres "github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

var _ = Describe("License", func() {
	Describe("Name/Fingerprint/ArchivePath/ArchiveDigest", func() {
		It("delegates to resource", func() {
			job := NewLicense(boshres.NewResourceWithBuiltArchive("name", "fp", "path", "sha1"))
			Expect(job.Name()).To(Equal("name"))
			Expect(job.Fingerprint()).To(Equal("fp"))
			Expect(job.ArchivePath()).To(Equal("path"))
			Expect(job.ArchiveDigest()).To(Equal("sha1"))
		})
	})
})
