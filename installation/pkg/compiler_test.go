package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_install_package "github.com/cloudfoundry/bosh-micro-cli/installation/pkg/mocks"
	mock_state_package "github.com/cloudfoundry/bosh-micro-cli/state/pkg/mocks"

	"errors"
	"fmt"
	"path"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	fakeblobstore "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
)

var _ = Describe("PackageCompiler", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		logger                  boshlog.Logger
		compiler                bmstatepkg.Compiler
		runner                  *fakesys.FakeCmdRunner
		pkg                     *bmrelpkg.Package
		fs                      *fakesys.FakeFileSystem
		compressor              *fakecmd.FakeCompressor
		packagesDir             string
		blobstore               *fakeblobstore.FakeBlobstore
		mockCompiledPackageRepo *mock_state_package.MockCompiledPackageRepo

		mockPackageInstaller *mock_install_package.MockInstaller

		dependency1 *bmrelpkg.Package
		dependency2 *bmrelpkg.Package
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		packagesDir = "fake-packages-dir"
		runner = fakesys.NewFakeCmdRunner()
		fs = fakesys.NewFakeFileSystem()
		compressor = fakecmd.NewFakeCompressor()

		mockPackageInstaller = mock_install_package.NewMockInstaller(mockCtrl)

		blobstore = fakeblobstore.NewFakeBlobstore()
		blobstore.CreateFingerprint = "fake-fingerprint"
		blobstore.CreateBlobID = "fake-blob-id"

		mockCompiledPackageRepo = mock_state_package.NewMockCompiledPackageRepo(mockCtrl)

		dependency1 = &bmrelpkg.Package{
			Name:        "fake-package-name-dependency-1",
			Fingerprint: "fake-package-fingerprint-dependency-1",
		}
		dependency2 = &bmrelpkg.Package{
			Name:        "fake-package-name-dependency-2",
			Fingerprint: "fake-package-fingerprint-dependency-2",
		}

		compiler = NewPackageCompiler(
			runner,
			packagesDir,
			fs,
			compressor,
			blobstore,
			mockCompiledPackageRepo,
			mockPackageInstaller,
			logger,
		)

		pkg = &bmrelpkg.Package{
			Name:          "fake-package-1",
			ExtractedPath: "/fake/path",
			Dependencies:  []*bmrelpkg.Package{dependency1, dependency2},
		}
	})

	Describe("Compile", func() {
		var (
			compiledPackageTarballPath string
			installPath                string

			expectPackageInstall1 *gomock.Call
			expectPackageInstall2 *gomock.Call
			expectFind            *gomock.Call
			expectSave            *gomock.Call
		)

		BeforeEach(func() {
			installPath = path.Join(packagesDir, pkg.Name)
			compiledPackageTarballPath = path.Join(packagesDir, "new-tarball.tgz")
		})

		JustBeforeEach(func() {
			expectFind = mockCompiledPackageRepo.EXPECT().Find(*pkg).Return(bmstatepkg.CompiledPackageRecord{}, false, nil).AnyTimes()

			compiledDependency1 := bmstatepkg.CompiledPackageRecord{
				BlobID:   "fake-dependency-blobstore-id-1",
				BlobSHA1: "fake-dependency-sha1-1",
			}
			mockCompiledPackageRepo.EXPECT().Find(*dependency1).Return(compiledDependency1, true, nil).AnyTimes()

			compiledPackageRef1 := CompiledPackageRef{
				Name:        "fake-package-name-dependency-1",
				Version:     "fake-package-fingerprint-dependency-1",
				BlobstoreID: "fake-dependency-blobstore-id-1",
				SHA1:        "fake-dependency-sha1-1",
			}
			expectPackageInstall1 = mockPackageInstaller.EXPECT().Install(compiledPackageRef1, packagesDir).AnyTimes()

			compiledDependency2 := bmstatepkg.CompiledPackageRecord{
				BlobID:   "fake-dependency-blobstore-id-2",
				BlobSHA1: "fake-dependency-sha1-2",
			}
			mockCompiledPackageRepo.EXPECT().Find(*dependency2).Return(compiledDependency2, true, nil).AnyTimes()

			compiledPackageRef2 := CompiledPackageRef{
				Name:        "fake-package-name-dependency-2",
				Version:     "fake-package-fingerprint-dependency-2",
				BlobstoreID: "fake-dependency-blobstore-id-2",
				SHA1:        "fake-dependency-sha1-2",
			}
			expectPackageInstall2 = mockPackageInstaller.EXPECT().Install(compiledPackageRef2, packagesDir).AnyTimes()

			// packaging file created when source is extracted
			fs.WriteFileString(path.Join(pkg.ExtractedPath, "packaging"), "")

			compressor.CompressFilesInDirTarballPath = compiledPackageTarballPath

			record := bmstatepkg.CompiledPackageRecord{
				BlobID:   "fake-blob-id",
				BlobSHA1: "fake-fingerprint",
			}
			expectSave = mockCompiledPackageRepo.EXPECT().Save(*pkg, record).AnyTimes()
		})

		Context("when the compiled package repo already has the package", func() {
			JustBeforeEach(func() {
				compiledPkgRecord := bmstatepkg.CompiledPackageRecord{
					BlobSHA1: "fake-fingerprint",
				}
				expectFind.Return(compiledPkgRecord, true, nil).Times(1)
			})

			It("skips the compilation", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(runner.RunComplexCommands)).To(Equal(0))
			})
		})

		It("installs all the dependencies for the package", func() {
			expectPackageInstall1.Times(1)
			expectPackageInstall2.Times(1)

			_, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())
		})

		It("runs the packaging script in package extractedPath dir", func() {
			_, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())

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
			_, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())

			Expect(compressor.CompressFilesInDirDir).To(Equal(installPath))
			Expect(compressor.CleanUpTarballPath).To(Equal(compiledPackageTarballPath))
		})

		It("moves the compressed package to a blobstore", func() {
			_, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())

			Expect(blobstore.CreateFileNames).To(Equal([]string{compiledPackageTarballPath}))
		})

		It("stores the compiled package blobID and fingerprint into the compile package repo", func() {
			expectSave.Times(1)

			_, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns the repo record", func() {
			record, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())

			Expect(record).To(Equal(bmstatepkg.CompiledPackageRecord{
				BlobID:   "fake-blob-id",
				BlobSHA1: "fake-fingerprint",
			}))
		})

		It("cleans up the packages dir", func() {
			_, err := compiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists(packagesDir)).To(BeFalse())
		})

		Context("when dependency installation fails", func() {
			JustBeforeEach(func() {
				expectPackageInstall1.Return(errors.New("fake-install-error")).Times(1)
			})

			It("returns an error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-install-error"))
			})
		})

		Context("when the packaging script does not exist", func() {
			JustBeforeEach(func() {
				err := fs.RemoveAll(path.Join(pkg.ExtractedPath, "packaging"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Packaging script for package 'fake-package-1' not found"))
			})
		})

		Context("when the packaging script fails", func() {
			JustBeforeEach(func() {
				fakeResult := fakesys.FakeCmdResult{
					ExitStatus: 1,
					Error:      errors.New("fake-error"),
				}
				runner.AddCmdResult("bash -x packaging", fakeResult)
			})

			It("returns error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Compiling package"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("when compression fails", func() {
			JustBeforeEach(func() {
				compressor.CompressFilesInDirErr = errors.New("fake-compression-error")
			})

			It("returns error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Compressing compiled package"))
			})
		})

		Context("when adding to blobstore fails", func() {
			JustBeforeEach(func() {
				blobstore.CreateErr = errors.New("fake-error")
			})

			It("returns error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Creating blob"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("when saving to the compiled package repo fails", func() {
			JustBeforeEach(func() {
				expectSave.Return(errors.New("fake-error")).Times(1)
			})

			It("returns error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Saving compiled package"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("when creating packages dir fails", func() {
			JustBeforeEach(func() {
				fs.RegisterMkdirAllError(installPath, errors.New("fake-error"))
			})

			It("returns error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Creating package install dir"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("when finding compiled package in the repo fails", func() {
			JustBeforeEach(func() {
				expectFind.Return(bmstatepkg.CompiledPackageRecord{}, false, errors.New("fake-error")).Times(1)
			})

			It("returns an error", func() {
				_, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Attempting to find compiled package '%s'", pkg.Name)))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
	})
})
