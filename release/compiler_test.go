package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("Compiler", func() {
	var (
		release  bmrelease.Release
		compiler bmrelease.Compiler
	)

	Context("Compile", func() {
		XContext("when the release is a valid", func() {
			It("compiles release without an error", func() {
				err := compiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
			})

			It("determines the order to compile packages", func() {

			})

			It("gets required package sources for each package in release", func() {

			})

			It("gets the dependencies for each package to compile", func() {

			})

			It("compiles each package", func() {

			})

			It("setup BOSH micro blobstore with entries for each compiled package", func() {

			})
		})

		Context("when the release has a bad package", func() {
			It("fails compilation for bad package", func() {

			})
		})
	})

	Context("DeterminePackageCompilationOrder", func() {
		var package1, package2 bmrelease.Package
		BeforeEach(func() {
			package1 = bmrelease.Package{
				Name: "fake-package-name-1",
			}
			package2 = bmrelease.Package{
				Name: "fake-package-name-2",
			}
			release = bmrelease.Release{
				Packages: []*bmrelease.Package{&package1, &package2},
			}
			compiler = bmrelease.NewCompiler()
		})
		Context("disjoint packages have a valid compilation sequence", func() {
			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := compiler.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(Equal([]*bmrelease.Package{&package1, &package2}))
			})
		})

		Context("dependent packages", func() {
			BeforeEach(func() {
				package1.Dependencies = []*bmrelease.Package{&package2}
			})

			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := compiler.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(Equal([]*bmrelease.Package{&package2, &package1}))
			})
		})

		Context("complex graph of dependent packages", func() {
			var package3, package4 bmrelease.Package
			BeforeEach(func() {
				package1.Dependencies = []*bmrelease.Package{&package2, &package3}
				package3 = bmrelease.Package{
					Name: "fake-package-name-3",
				}
				package4 = bmrelease.Package{
					Name:         "fake-package-name-4",
					Dependencies: []*bmrelease.Package{&package3, &package2},
				}
				release.Packages = append(release.Packages, &package3, &package4)
			})

			It("returns an ordered set of package compilation", func() {
				compilationOrder, err := compiler.DeterminePackageCompilationOrder(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(compilationOrder).To(Equal([]*bmrelease.Package{&package2, &package3, &package1, &package4}))
			})
		})

		Context("packages have a invalid compilation sequence", func() {
			It("fails with error", func() {

			})
		})
	})
})
