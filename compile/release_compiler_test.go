package compile_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakeboshcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakebmreal "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/compile"
)

var _ = Describe("ReleaseCompiler", func() {
	var (
		release             bmrel.Release
		releaseCompiler     ReleaseCompiler
		fakeDA              *fakebmreal.FakeDependencyAnalysis
		fakePackageCompiler *fakeboshcomp.FakePackageCompiler
	)

	BeforeEach(func() {
		fakeDA = fakebmreal.NewFakeDependencyAnalysis()
		fakePackageCompiler = fakeboshcomp.NewFakePackageCompiler()

		releaseCompiler = NewReleaseCompiler(fakeDA, fakePackageCompiler)
		release = bmrel.Release{}
	})

	Context("Compile", func() {
		Context("when the release", func() {
			var expectedPackages []*bmrel.Package

			BeforeEach(func() {
				package1 := bmrel.Package{Name: "fake-package-1"}
				package2 := bmrel.Package{Name: "fake-package-2"}

				expectedPackages = []*bmrel.Package{&package1, &package2}

				fakeDA.DeterminePackageCompilationOrderResult = []*bmrel.Package{
					&package1,
					&package2,
				}
			})

			It("determines the order to compile packages", func() {
				err := releaseCompiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeDA.DeterminePackageCompilationOrderRelease).To(Equal(release))
			})

			It("compiles each package", func() {
				err := releaseCompiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakePackageCompiler.CompilePackages).To(Equal(expectedPackages))
			})

			It("compiles each package and returns error for first package", func() {
				fakePackageCompiler.CompileError = errors.New("Compilation failed")
				err := releaseCompiler.Compile(release)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Package `fake-package-1' compilation failed"))
			})

			It("stops compiling after the first failures", func() {
				fakePackageCompiler.CompileError = errors.New("Compilation failed")
				err := releaseCompiler.Compile(release)
				Expect(err).To(HaveOccurred())
				Expect(len(fakePackageCompiler.CompilePackages)).To(Equal(1))
			})
		})
	})
})
