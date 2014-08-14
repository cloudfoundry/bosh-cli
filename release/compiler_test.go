package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = XDescribe("Compiler", func() {
	var (
		release  bmrelease.Release
		compiler bmrelease.Compiler
	)

	Context("Compile", func() {
		Context("when the release is a valid", func() {
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
})
