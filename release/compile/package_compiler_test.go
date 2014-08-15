package compile_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshsys "github.com/cloudfoundry/bosh-agent/system"

	fakecmdrunner "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	. "github.com/cloudfoundry/bosh-micro-cli/release/compile"
)

var _ = Describe("PackageCompiler", func() {
	var (
		pc     PackageCompiler
		runner *fakecmdrunner.FakeCmdRunner
		pkg    *bmrel.Package
	)

	BeforeEach(func() {

		runner = fakecmdrunner.NewFakeCmdRunner()
		pc = NewPackageCompiler(runner)
		pkg = &bmrel.Package{
			Name:          "fake-package-1",
			Version:       "fake-package-version",
			ExtractedPath: "/fake/path",
		}
	})

	Describe("compiling a package", func() {
		BeforeEach(func() {
			err := pc.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when compilcation succeeds", func() {
			It("runs the packaging script in package extractedPath dir", func() {
				expectedCmd := boshsys.Command{
					Name: "bash",
					Args: []string{"-x", "packaging"},
					Env: map[string]string{
						"BOSH_COMPILE_TARGET":  pkg.ExtractedPath,
						"BOSH_INSTALL_TARGET":  "/fake-dir/packages/pkg_name",
						"BOSH_PACKAGE_NAME":    pkg.Name,
						"BOSH_PACKAGE_VERSION": pkg.Version,
						"BOSH_PACKAGES_DIR":    "/fake-packages-dir/",
					},
					WorkingDir: pkg.ExtractedPath,
				}

				Expect(runner.RunComplexCommands).To(HaveLen(1))
				Expect(runner.RunComplexCommands[0]).To(Equal(expectedCmd))
			})
			XIt("store the compiled package to a storage", func() {})
			XIt("cleans up the working dir")
		})

		Context("when compilcation fails", func() {
			XIt("returns error to the caller")
			XIt("cleans up the working dir even when the compilation fails")
		})
	})
})
