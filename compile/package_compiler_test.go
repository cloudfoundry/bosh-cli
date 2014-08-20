package compile_test

import (
	"errors"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeblobstore "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeboshcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
	. "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("PackageCompiler", func() {
	var (
		pc                  PackageCompiler
		runner              *fakesys.FakeCmdRunner
		pkg                 *bmrel.Package
		fs                  *fakesys.FakeFileSystem
		compressor          *fakecmd.FakeCompressor
		packagesDir         string
		blobstore           *fakeblobstore.FakeBlobstore
		compiledPackageRepo *fakeboshcomp.FakeCompiledPackageRepo
	)

	BeforeEach(func() {
		packagesDir = "fake-packages-dir"
		runner = fakesys.NewFakeCmdRunner()
		fs = fakesys.NewFakeFileSystem()
		compressor = fakecmd.NewFakeCompressor()

		blobstore = fakeblobstore.NewFakeBlobstore()
		blobstore.CreateFingerprint = "fake-fingerprint"
		blobstore.CreateBlobID = "fake-blob-id"

		compiledPackageRepo = fakeboshcomp.NewFakeCompiledPackageRepo()

		pc = NewPackageCompiler(runner, packagesDir, fs, compressor, blobstore, compiledPackageRepo)
		pkg = &bmrel.Package{
			Name:          "fake-package-1",
			Version:       "fake-package-version",
			ExtractedPath: "/fake/path",
		}
	})

	Describe("compiling a package", func() {
		var newTarballPath string
		var installPath string
		Context("when the packing script exists", func() {
			BeforeEach(func() {
				installPath = path.Join(packagesDir, pkg.Name)
				fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
				newTarballPath = path.Join(packagesDir, "new-tarball")
				compressor.CompressFilesInDirTarballPath = newTarballPath
				err := pc.Compile(pkg)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when compilation succeeds", func() {
				It("runs the packaging script in package extractedPath dir", func() {
					expectedCmd := boshsys.Command{
						Name: "bash",
						Args: []string{"-x", "packaging"},
						Env: map[string]string{
							"BOSH_COMPILE_TARGET":  pkg.ExtractedPath,
							"BOSH_INSTALL_TARGET":  installPath,
							"BOSH_PACKAGE_NAME":    pkg.Name,
							"BOSH_PACKAGE_VERSION": pkg.Version,
							"BOSH_PACKAGES_DIR":    packagesDir,
						},
						WorkingDir: pkg.ExtractedPath,
					}

					Expect(runner.RunComplexCommands).To(HaveLen(1))
					Expect(runner.RunComplexCommands[0]).To(Equal(expectedCmd))
				})

				It("compresses the compiled package", func() {
					Expect(compressor.CompressFilesInDirDir).To(Equal(installPath))
					Expect(compressor.CleanUpTarballPath).To(Equal(newTarballPath))
				})

				It("moves the compressed package to a blobstore", func() {
					Expect(blobstore.CreateFileName).To(Equal(newTarballPath))
				})

				It("stores the compiled package blobID and fingerprint into the compile package repo", func() {
					Expect(compiledPackageRepo.SavePackage).To(Equal(*pkg))
					Expect(compiledPackageRepo.SaveRecord).To(Equal(CompiledPackageRecord{
						"fake-blob-id",
						"fake-fingerprint",
					}))
				})

				It("cleans up the packages dir", func() {
					Expect(fs.FileExists(packagesDir)).To(BeFalse())
				})
			})
		})

		Describe("compilation failures", func() {
			Context("when the packaging script does not exist", func() {
				It("returns error", func() {
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Packaging script for package `fake-package-1' not found"))
				})
			})

			Context("when the packaging script fails", func() {
				It("returns error", func() {
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					fakeResult := fakesys.FakeCmdResult{
						ExitStatus: 1,
						Error:      errors.New("fake-error"),
					}
					runner.AddCmdResult("bash -x packaging", fakeResult)
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Compiling package"))
				})
			})

			Context("when compression fails", func() {
				It("returns error", func() {
					compressor.CompressFilesInDirErr = errors.New("fake-compression-error")
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Compressing compiled package"))
				})
			})

			Context("when adding to blobstore fails", func() {
				It("returns error", func() {
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					blobstore.CreateErr = errors.New("fake-create-err")
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Creating blob"))
				})
			})

			Context("when saving to the compiled package repo fails", func() {
				It("returns error", func() {
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					compiledPackageRepo.SaveError = errors.New("fake-save-compiled-package-error")
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Saving compiled package"))
				})
			})

			Context("when creating packages dir fails", func() {
				It("returns error", func() {
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					fs.RegisterMkdirAllError(installPath, errors.New("fake-mkdir-error"))
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Creating package install dir"))
				})
			})

			It("cleans up the working dir", func() {
				err := pc.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(fs.FileExists(packagesDir)).To(BeFalse())
			})
		})
	})
})
