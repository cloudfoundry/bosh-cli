package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/release/pkg"
)

var _ = Describe("NewCompiledPackageWithoutArchive", func() {
	var (
		compiledPkg *CompiledPackage
	)

	BeforeEach(func() {
		compiledPkg = NewCompiledPackageWithoutArchive(
			"name", "fp", "os-slug", "sha1", []string{"pkg1", "pkg2"})
	})

	Describe("common methods", func() {
		It("returns values", func() {
			Expect(compiledPkg.Name()).To(Equal("name"))
			Expect(compiledPkg.Fingerprint()).To(Equal("fp"))
			Expect(compiledPkg.OSVersionSlug()).To(Equal("os-slug"))

			Expect(func() { compiledPkg.ArchivePath() }).To(Panic())
			Expect(compiledPkg.ArchiveSHA1()).To(Equal("sha1"))

			Expect(compiledPkg.DependencyNames()).To(Equal([]string{"pkg1", "pkg2"}))
		})
	})

	Describe("AttachDependencies", func() {
		It("attaches dependencies based on their names", func() {
			pkg1 := NewCompiledPackageWithoutArchive("pkg1", "fp", "os-slug", "sha1", nil)
			pkg2 := NewCompiledPackageWithoutArchive("pkg2", "fp", "os-slug", "sha1", nil)
			unusedPkg := NewCompiledPackageWithoutArchive("unused", "fp", "os-slug", "sha1", nil)

			err := compiledPkg.AttachDependencies([]*CompiledPackage{pkg1, unusedPkg, pkg2})
			Expect(err).ToNot(HaveOccurred())

			Expect(compiledPkg.Dependencies).To(Equal([]*CompiledPackage{pkg1, pkg2}))
		})

		It("returns error if dependency cannot be found", func() {
			pkg2 := NewCompiledPackageWithoutArchive("pkg2", "fp", "os-slug", "sha1", nil)

			err := compiledPkg.AttachDependencies([]*CompiledPackage{pkg2})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find compiled package 'pkg1' since it's a dependency of compiled package 'name'"))
		})
	})
})

var _ = Describe("NewCompiledPackageWithArchive", func() {
	var (
		compiledPkg *CompiledPackage
	)

	BeforeEach(func() {
		compiledPkg = NewCompiledPackageWithArchive(
			"name", "fp", "os-slug", "path", "sha1", []string{"pkg1", "pkg2"})
	})

	Describe("common methods", func() {
		It("returns values", func() {
			Expect(compiledPkg.Name()).To(Equal("name"))
			Expect(compiledPkg.Fingerprint()).To(Equal("fp"))
			Expect(compiledPkg.OSVersionSlug()).To(Equal("os-slug"))

			Expect(compiledPkg.ArchivePath()).To(Equal("path"))
			Expect(compiledPkg.ArchiveSHA1()).To(Equal("sha1"))

			Expect(compiledPkg.DependencyNames()).To(Equal([]string{"pkg1", "pkg2"}))
		})
	})

	Describe("AttachDependencies", func() {
		It("attaches dependencies based on their names", func() {
			pkg1 := NewCompiledPackageWithArchive("pkg1", "fp", "os-slug", "path", "sha1", nil)
			pkg2 := NewCompiledPackageWithArchive("pkg2", "fp", "os-slug", "path", "sha1", nil)
			unusedPkg := NewCompiledPackageWithArchive("unused", "fp", "os-slug", "path", "sha1", nil)

			err := compiledPkg.AttachDependencies([]*CompiledPackage{pkg1, unusedPkg, pkg2})
			Expect(err).ToNot(HaveOccurred())

			Expect(compiledPkg.Dependencies).To(Equal([]*CompiledPackage{pkg1, pkg2}))
		})

		It("returns error if dependency cannot be found", func() {
			pkg2 := NewCompiledPackageWithArchive("pkg2", "fp", "os-slug", "path", "sha1", nil)

			err := compiledPkg.AttachDependencies([]*CompiledPackage{pkg2})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find compiled package 'pkg1' since it's a dependency of compiled package 'name'"))
		})
	})
})
