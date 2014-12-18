package compile_test

import (
	"errors"
	"fmt"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshsys "github.com/cloudfoundry/bosh-agent/system"

	fakeblobstore "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	fakebmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install/fakes"
	fakebmpkgs "github.com/cloudfoundry/bosh-micro-cli/cpi/packages/fakes"

	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/cpi/packages"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/cpi/compile"
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
		compiledPackageRepo *fakebmpkgs.FakeCompiledPackageRepo
		packageInstaller    *fakebmcpiinstall.FakePackageInstaller
		dependency1         *bmrel.Package
		dependency2         *bmrel.Package
	)

	BeforeEach(func() {
		packagesDir = "fake-packages-dir"
		runner = fakesys.NewFakeCmdRunner()
		fs = fakesys.NewFakeFileSystem()
		compressor = fakecmd.NewFakeCompressor()
		packageInstaller = fakebmcpiinstall.NewFakePackageInstaller()

		blobstore = fakeblobstore.NewFakeBlobstore()
		blobstore.CreateFingerprint = "fake-fingerprint"
		blobstore.CreateBlobID = "fake-blob-id"

		compiledPackageRepo = fakebmpkgs.NewFakeCompiledPackageRepo()

		dependency1 = &bmrel.Package{
			Name: "fake-dependency-1",
		}
		dependency2 = &bmrel.Package{
			Name: "fake-dependency-1",
		}

		pc = NewPackageCompiler(
			runner,
			packagesDir,
			fs,
			compressor,
			blobstore,
			compiledPackageRepo,
			packageInstaller,
		)
		pkg = &bmrel.Package{
			Name:          "fake-package-1",
			ExtractedPath: "/fake/path",
			Dependencies:  []*bmrel.Package{dependency1, dependency2},
		}
	})

	Describe("compiling a package", func() {
		var newTarballPath string
		var installPath string

		BeforeEach(func() {
			compiledPackageRepo.SetFindBehavior(*pkg, bmpkgs.CompiledPackageRecord{}, false, nil)
			compiledPackageRepo.SetFindBehavior(*dependency1, bmpkgs.CompiledPackageRecord{}, false, nil)
			compiledPackageRepo.SetFindBehavior(*dependency2, bmpkgs.CompiledPackageRecord{}, false, nil)

			packageInstaller.SetInstallBehavior(dependency1, packagesDir, nil)
			packageInstaller.SetInstallBehavior(dependency2, packagesDir, nil)
		})

		Context("when the compiled package repo already has the package", func() {
			BeforeEach(func() {
				compiledPkgRecord := bmpkgs.CompiledPackageRecord{
					BlobSHA1: "fake-fingerprint",
				}
				compiledPackageRepo.SetFindBehavior(*pkg, compiledPkgRecord, true, nil)
				fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")

				err := pc.Compile(pkg)
				Expect(err).ToNot(HaveOccurred())
			})

			It("skips the compilation", func() {
				Expect(len(runner.RunComplexCommands)).To(Equal(0))
			})
		})

		Context("when compilation succeeds", func() {
			BeforeEach(func() {
				installPath = path.Join(packagesDir, pkg.Name)
				fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
				newTarballPath = path.Join(packagesDir, "new-tarball")
				compressor.CompressFilesInDirTarballPath = newTarballPath

				record := bmpkgs.CompiledPackageRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-fingerprint",
				}
				compiledPackageRepo.SetSaveBehavior(*pkg, record, nil)

				err := pc.Compile(pkg)
				Expect(err).ToNot(HaveOccurred())
			})

			It("installs all the dependencies for the package", func() {
				Expect(packageInstaller.InstallInputs).To(ContainElement(
					fakebmcpiinstall.InstallInput{
						Package: dependency1,
						Target:  path.Join(packagesDir),
					},
				))
				Expect(packageInstaller.InstallInputs).To(ContainElement(
					fakebmcpiinstall.InstallInput{
						Package: dependency2,
						Target:  path.Join(packagesDir),
					},
				))
			})

			It("runs the packaging script in package extractedPath dir", func() {
				expectedCmd := boshsys.Command{
					Name: "bash",
					Args: []string{"-x", "packaging"},
					Env: map[string]string{
						"BOSH_COMPILE_TARGET": pkg.ExtractedPath,
						"BOSH_INSTALL_TARGET": installPath,
						"BOSH_PACKAGE_NAME":   pkg.Name,
						"BOSH_PACKAGES_DIR":   packagesDir,
						"PATH":                "/usr/local/bin:/usr/bin:/bin",
					},
					UseIsolatedEnv: true,
					WorkingDir:     pkg.ExtractedPath,
				}

				Expect(runner.RunComplexCommands).To(HaveLen(1))
				Expect(runner.RunComplexCommands[0]).To(Equal(expectedCmd))
			})

			It("compresses the compiled package", func() {
				Expect(compressor.CompressFilesInDirDir).To(Equal(installPath))
				Expect(compressor.CleanUpTarballPath).To(Equal(newTarballPath))
			})

			It("moves the compressed package to a blobstore", func() {
				Expect(blobstore.CreateFileNames).To(Equal([]string{newTarballPath}))
			})

			It("stores the compiled package blobID and fingerprint into the compile package repo", func() {
				record := bmpkgs.CompiledPackageRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-fingerprint",
				}
				Expect(compiledPackageRepo.SaveInputs).To(ContainElement(
					fakebmpkgs.SaveInput{Package: *pkg, Record: record},
				))
			})

			It("cleans up the packages dir", func() {
				Expect(fs.FileExists(packagesDir)).To(BeFalse())
			})
		})

		Context("when compilation fails", func() {
			Context("when depedency installation fails", func() {
				BeforeEach(func() {
					packageInstaller.SetInstallBehavior(
						dependency1,
						path.Join(packagesDir),
						errors.New("fake-error"),
					)
				})

				It("returns an error", func() {
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Installing package"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			Context("when the packaging script does not exist", func() {
				It("returns error", func() {
					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Packaging script for package 'fake-package-1' not found"))
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
					Expect(err.Error()).To(ContainSubstring("fake-error"))
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
					blobstore.CreateErr = errors.New("fake-error")

					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Creating blob"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			Context("when saving to the compiled package repo fails", func() {
				It("returns error", func() {
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					record := bmpkgs.CompiledPackageRecord{
						BlobID:   "fake-blob-id",
						BlobSHA1: "fake-fingerprint",
					}
					compiledPackageRepo.SetSaveBehavior(*pkg, record, errors.New("fake-error"))

					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Saving compiled package"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			Context("when creating packages dir fails", func() {
				It("returns error", func() {
					fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")
					fs.RegisterMkdirAllError(installPath, errors.New("fake-error"))

					err := pc.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Creating package install dir"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			It("cleans up the working dir", func() {
				err := pc.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(fs.FileExists(packagesDir)).To(BeFalse())
			})
		})

		It("errors when finding package from compiled package repo fails", func() {
			compiledPackageRepo.SetFindBehavior(*pkg, bmpkgs.CompiledPackageRecord{}, false, errors.New("fake-error"))

			err := pc.Compile(pkg)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Attempting to find compiled package '%s'", pkg.Name)))
			Expect(err.Error()).To(ContainSubstring("fake-error"))
		})
	})
})
