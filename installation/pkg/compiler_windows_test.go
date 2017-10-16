package pkg_test

import (
	"errors"
	"path/filepath"

	fakeblobstore "github.com/cloudfoundry/bosh-utils/blobstore/fakes"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	fakecmd "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/installation/blobextract/fakeblobextract"
	. "github.com/cloudfoundry/bosh-cli/installation/pkg"
	birelpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
	bistatepkg "github.com/cloudfoundry/bosh-cli/state/pkg"
	mock_state_package "github.com/cloudfoundry/bosh-cli/state/pkg/mocks"
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
		compiler                bistatepkg.Compiler
		runner                  *fakesys.FakeCmdRunner
		pkg                     *birelpkg.Package
		fs                      *fakesys.FakeFileSystem
		compressor              *fakecmd.FakeCompressor
		packagesDir             string
		blobstore               *fakeblobstore.FakeDigestBlobstore
		mockCompiledPackageRepo *mock_state_package.MockCompiledPackageRepo

		fakeExtractor *fakeblobextract.FakeExtractor

		dependency1 *birelpkg.Package
		dependency2 *birelpkg.Package
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		packagesDir = "fake-packages-dir"
		runner = fakesys.NewFakeCmdRunner()
		fs = fakesys.NewFakeFileSystem()
		compressor = fakecmd.NewFakeCompressor()

		fakeExtractor = &fakeblobextract.FakeExtractor{}

		blobstore = &fakeblobstore.FakeDigestBlobstore{}
		digest := boshcrypto.MustParseMultipleDigest("fakefingerprint")

		blobstore.CreateReturns("fake-blob-id", digest, nil)

		mockCompiledPackageRepo = mock_state_package.NewMockCompiledPackageRepo(mockCtrl)

		dependency1 = birelpkg.NewPackage(NewResource("pkg-dep1-name", "", nil), nil)
		dependency2 = birelpkg.NewPackage(NewResource("pkg-dep2-name", "", nil), nil)
		pkg = birelpkg.NewExtractedPackage(NewResource("pkg1-name", "", nil), []string{"pkg-dep1-name", "pkg-dep2-name"}, filepath.Join("/", "pkg-dir"), fs)
		pkg.AttachDependencies([]*birelpkg.Package{dependency1, dependency2})

		compiler = NewPackageCompiler(
			runner,
			packagesDir,
			fs,
			compressor,
			blobstore,
			mockCompiledPackageRepo,
			fakeExtractor,
			logger,
		)
	})

	Describe("Compile", func() {
		var (
			compiledPackageTarballPath string
			installPath                string

			dep1 bistatepkg.CompiledPackageRecord
			dep2 bistatepkg.CompiledPackageRecord

			expectFind *gomock.Call
			expectSave *gomock.Call
		)

		BeforeEach(func() {
			installPath = filepath.Join(packagesDir, "pkg1-name")
			compiledPackageTarballPath = filepath.Join(packagesDir, "new-tarball.tgz")
		})

		JustBeforeEach(func() {
			expectFind = mockCompiledPackageRepo.EXPECT().Find(pkg).Return(bistatepkg.CompiledPackageRecord{}, false, nil).AnyTimes()

			dep1 = bistatepkg.CompiledPackageRecord{
				BlobID:   "fake-dependency-blobstore-id-1",
				BlobSHA1: "fake-dependency-sha1-1",
			}
			mockCompiledPackageRepo.EXPECT().Find(dependency1).Return(dep1, true, nil).AnyTimes()

			dep2 = bistatepkg.CompiledPackageRecord{
				BlobID:   "fake-dependency-blobstore-id-2",
				BlobSHA1: "fake-dependency-sha1-2",
			}
			mockCompiledPackageRepo.EXPECT().Find(dependency2).Return(dep2, true, nil).AnyTimes()

			// packaging file created when source is extracted
			fs.WriteFileString(filepath.Join("/", "pkg-dir", "packaging"), "")

			compressor.CompressFilesInDirTarballPath = compiledPackageTarballPath

			record := bistatepkg.CompiledPackageRecord{
				BlobID:   "fake-blob-id",
				BlobSHA1: "fakefingerprint",
			}
			expectSave = mockCompiledPackageRepo.EXPECT().Save(pkg, record).AnyTimes()
		})

		Context("when obtaining working directory fails", func() {
			JustBeforeEach(func() {
				fakeResult := fakesys.FakeCmdResult{
					ExitStatus: 1,
					Error:      errors.New("fake-error"),
				}
				runner.AddCmdResult("bash -c pwd", fakeResult)
			})

			It("returns error", func() {
				_, _, err := compiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Obtaining working directory"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
	})
})
