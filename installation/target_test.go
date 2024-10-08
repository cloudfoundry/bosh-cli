package installation_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/installation"
)

var _ = Describe("Target", func() {
	Describe("Paths", func() {
		var target Target
		BeforeEach(func() {
			target = NewTarget("/home/fake/madcow", "")
		})

		It("returns the blobstore path", func() {
			Expect(target.BlobstorePath()).To(Equal(filepath.Join("/", "home", "fake", "madcow", "blobs")))
		})

		It("returns the compiled packages index path", func() {
			Expect(target.CompiledPackagedIndexPath()).To(Equal(filepath.Join("/", "home", "fake", "madcow", "compiled_packages.json")))
		})

		It("returns the templates index path", func() {
			Expect(target.TemplatesIndexPath()).To(Equal(filepath.Join("/", "home", "fake", "madcow", "templates.json")))
		})

		Context("packageDir is NOT provided", func() {
			It("returns the packages path as a subdirectory of the path", func() {
				Expect(target.PackagesPath()).To(Equal(filepath.Join("/", "home", "fake", "madcow", "packages")))
			})
		})

		Context("packageDir is provided", func() {
			var packagesDir string
			BeforeEach(func() {
				packagesDir = "/some/good/path"
				target = NewTarget("/home/fake/madcow", packagesDir)
			})

			It("returns the provided packages path", func() {
				Expect(target.PackagesPath()).To(Equal(packagesDir))
			})
		})

		It("returns the temp path", func() {
			Expect(target.TmpPath()).To(Equal(filepath.Join("/", "home", "fake", "madcow", "tmp")))
		})
	})
})
