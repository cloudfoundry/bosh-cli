package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeboshcomp "github.com/cloudfoundry/bosh-agent/agent/compiler/fakes"

	fakebmreal "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("Compiler", func() {
	var (
		release             Release
		compiler            Compiler
		fakeDA              *fakebmreal.FakeDependencyAnalysis
		fakePackageCompiler *fakeboshcomp.FakeCompiler
	)

	Context("Compile", func() {
		BeforeEach(func() {
			fakeDA = fakebmreal.NewFakeDependencyAnalysis()

			fakePackageCompiler = fakeboshcomp.NewFakeCompiler()

			compiler = NewCompiler(fakeDA, fakePackageCompiler)
			release = Release{}
		})

		Context("when the release is a valid", func() {
			It("compiles release without an error", func() {
				err := compiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
			})

			It("determines the order to compile packages", func() {
				err := compiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeDA.DeterminePackageCompilationOrderRelease).To(Equal(release))
			})

			It("compiles each package", func() {
			})

			//It("setup BOSH micro blobstore with entries for each compiled package", func() {
			//})
		})

		Context("when the release has a bad package", func() {
			It("fails compilation for bad package", func() {

			})
		})
	})
})
