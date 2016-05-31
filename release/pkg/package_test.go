package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/release/pkg"
	boshres "github.com/cloudfoundry/bosh-init/release/resource"
)

var _ = Describe("Package", func() {
	Describe("common methods", func() {
		It("delegates to resource", func() {
			pkg := NewPackage(boshres.NewResourceWithBuiltArchive("name", "fp", "path", "sha1"), []string{"pkg1"})
			Expect(pkg.Name()).To(Equal("name"))
			Expect(pkg.String()).To(Equal("name"))
			Expect(pkg.Fingerprint()).To(Equal("fp"))
			Expect(pkg.ArchivePath()).To(Equal("path"))
			Expect(pkg.ArchiveSHA1()).To(Equal("sha1"))
			Expect(pkg.DependencyNames()).To(Equal([]string{"pkg1"}))
		})
	})

	Describe("AttachDependencies", func() {
		It("attaches dependencies based on their names", func() {
			pkg := NewPackage(boshres.NewResourceWithBuiltArchive("name", "fp", "path", "sha1"), []string{"pkg1", "pkg2"})
			pkg1 := NewPackage(boshres.NewResourceWithBuiltArchive("pkg1", "fp", "path", "sha1"), nil)
			pkg2 := NewPackage(boshres.NewResourceWithBuiltArchive("pkg2", "fp", "path", "sha1"), nil)
			unusedPkg := NewPackage(boshres.NewResourceWithBuiltArchive("unused", "fp", "path", "sha1"), nil)

			err := pkg.AttachDependencies([]*Package{pkg1, unusedPkg, pkg2})
			Expect(err).ToNot(HaveOccurred())

			Expect(pkg.Dependencies).To(Equal([]*Package{pkg1, pkg2}))
		})

		It("returns error if dependency cannot be found", func() {
			pkg := NewPackage(boshres.NewResourceWithBuiltArchive("name", "fp", "path", "sha1"), []string{"pkg1"})
			pkg2 := NewPackage(boshres.NewResourceWithBuiltArchive("pkg2", "fp", "path", "sha1"), nil)

			err := pkg.AttachDependencies([]*Package{pkg2})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find package 'pkg1' since it's a dependency of package 'name'"))
		})
	})

	Describe("CleanUp", func() {
		It("does nothing by default", func() {
			pkg := NewPackage(boshres.NewResourceWithBuiltArchive("name", "fp", "path", "sha1"), nil)
			Expect(pkg.CleanUp()).ToNot(HaveOccurred())
		})
	})
})
