package state_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"

	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
)

var _ = Describe("RemotePackageCompiler", describeRemotePackageCompiler)

func describeRemotePackageCompiler() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		packageRepo bmstatepkg.CompiledPackageRepo

		pkgDependency    *bmrelpkg.Package
		pkg              *bmrelpkg.Package
		pkgDependencyRef PackageRef

		mockBlobstore   *mock_blobstore.MockBlobstore
		mockAgentClient *mock_agentclient.MockAgentClient

		archivePath = "fake-archive-path"

		remotePackageCompiler bmstatepkg.Compiler

		compiledPackages map[bmstatepkg.CompiledPackageRecord]*bmrelpkg.Package

		expectBlobstoreAdd *gomock.Call
		expectAgentCompile *gomock.Call
	)

	BeforeEach(func() {
		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)
		mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)

		index := bmindex.NewInMemoryIndex()
		packageRepo = bmstatepkg.NewCompiledPackageRepo(index)
		remotePackageCompiler = NewRemotePackageCompiler(mockBlobstore, mockAgentClient, packageRepo)

		pkgDependency = &bmrelpkg.Package{
			Name:        "fake-package-name-dep",
			Fingerprint: "fake-package-fingerprint-dep",
		}

		pkg = &bmrelpkg.Package{
			Name:         "fake-package-name",
			Fingerprint:  "fake-package-fingerprint",
			SHA1:         "fake-source-package-sha1",
			ArchivePath:  archivePath,
			Dependencies: []*bmrelpkg.Package{pkgDependency},
		}

		pkgDependencyRef = PackageRef{
			Name:    "fake-package-name-dep",
			Version: "fake-package-fingerprint-dep",
			Archive: BlobRef{
				BlobstoreID: "fake-compiled-package-blob-id-dep",
				SHA1:        "fake-compiled-package-sha1-dep",
			},
		}

		depRecord1 := bmstatepkg.CompiledPackageRecord{
			BlobID:   "fake-compiled-package-blob-id-dep",
			BlobSHA1: "fake-compiled-package-sha1-dep",
		}

		compiledPackages = map[bmstatepkg.CompiledPackageRecord]*bmrelpkg.Package{
			depRecord1: pkgDependency,
		}
	})

	JustBeforeEach(func() {
		// add compiled packages to the repo
		for record, dependency := range compiledPackages {
			err := packageRepo.Save(*dependency, record)
			Expect(err).ToNot(HaveOccurred())
		}

		packageSource := bmagentclient.BlobRef{
			Name:        "fake-package-name",
			Version:     "fake-package-fingerprint",
			BlobstoreID: "fake-source-package-blob-id",
			SHA1:        "fake-source-package-sha1",
		}
		packageDependencies := []bmagentclient.BlobRef{
			{
				Name:        "fake-package-name-dep",
				Version:     "fake-package-fingerprint-dep",
				BlobstoreID: "fake-compiled-package-blob-id-dep",
				SHA1:        "fake-compiled-package-sha1-dep",
			},
		}
		compiledPackageRef := bmagentclient.BlobRef{
			Name:        "fake-package-name",
			Version:     "fake-package-version",
			BlobstoreID: "fake-compiled-package-blob-id",
			SHA1:        "fake-compiled-package-sha1",
		}

		expectBlobstoreAdd = mockBlobstore.EXPECT().Add(archivePath).Return("fake-source-package-blob-id", nil).AnyTimes()
		expectAgentCompile = mockAgentClient.EXPECT().CompilePackage(packageSource, packageDependencies).Return(compiledPackageRef, nil).AnyTimes()
	})

	Describe("Compile", func() {
		It("uploads the package archive to the blobstore and then compiles the package with the agent", func() {
			gomock.InOrder(
				expectBlobstoreAdd.Times(1),
				expectAgentCompile.Times(1),
			)

			compiledPackageRecord, err := remotePackageCompiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(compiledPackageRecord).To(Equal(bmstatepkg.CompiledPackageRecord{
				BlobID:   "fake-compiled-package-blob-id",
				BlobSHA1: "fake-compiled-package-sha1",
			}))
		})

		It("saves the compiled package ref in the package repo", func() {
			compiledPackageRecord, err := remotePackageCompiler.Compile(pkg)
			Expect(err).ToNot(HaveOccurred())

			record, found, err := packageRepo.Find(*pkg)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(record).To(Equal(compiledPackageRecord))
		})

		Context("when the dependencies are not in the repo", func() {
			BeforeEach(func() {
				compiledPackages = map[bmstatepkg.CompiledPackageRecord]*bmrelpkg.Package{}
			})

			It("returns an error", func() {
				_, err := remotePackageCompiler.Compile(pkg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Remote compilation failure: Package 'fake-package-name' requires package 'fake-package-name-dep', but it has not been compiled"))
			})
		})
	})
}
