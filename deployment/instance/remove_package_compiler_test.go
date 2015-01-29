package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"

	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("RemotePackageCompiler", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		pkg              *bmrel.Package
		pkgDependencyRef PackageRef

		mockBlobstore   *mock_blobstore.MockBlobstore
		mockAgentClient *mock_agentclient.MockAgentClient

		archivePath = "fake-archive-path"

		remotePackageCompiler PackageCompiler
	)

	BeforeEach(func() {
		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)
		mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)
		remotePackageCompiler = NewRemotePackageCompiler(mockBlobstore, mockAgentClient)

		pkgDependency := &bmrel.Package{
			Name:        "fake-package-name-dep",
			Fingerprint: "fake-package-fingerprint-dep",
		}

		pkg = &bmrel.Package{
			Name:         "fake-package-name",
			Fingerprint:  "fake-package-fingerprint",
			SHA1:         "fake-source-package-sha1",
			ArchivePath:  archivePath,
			Dependencies: []*bmrel.Package{pkgDependency},
		}

		pkgDependencyRef = PackageRef{
			Name:    "fake-package-name-dep",
			Version: "fake-package-fingerprint-dep",
			Archive: BlobRef{
				BlobstoreID: "fake-compiled-package-blob-id-dep",
				SHA1:        "fake-compiled-package-sha1-dep",
			},
		}
	})

	JustBeforeEach(func() {
		packageSource := bmagentclient.BlobRef{
			Name:        "fake-package-name",
			Version:     "fake-package-fingerprint",
			SHA1:        "fake-source-package-sha1",
			BlobstoreID: "fake-source-package-blob-id",
		}
		packageDependencies := []bmagentclient.BlobRef{
			{
				Name:        "fake-package-name-dep",
				Version:     "fake-package-fingerprint-dep",
				SHA1:        "fake-compiled-package-sha1-dep",
				BlobstoreID: "fake-compiled-package-blob-id-dep",
			},
		}
		compiledPackageRef := bmagentclient.BlobRef{
			Name:        "fake-package-name",
			Version:     "fake-package-version",
			SHA1:        "fake-compiled-package-sha1",
			BlobstoreID: "fake-compiled-package-blob-id",
		}

		gomock.InOrder(
			mockBlobstore.EXPECT().Add(archivePath).Return("fake-source-package-blob-id", nil),
			mockAgentClient.EXPECT().CompilePackage(packageSource, packageDependencies).Return(compiledPackageRef, nil),
		)
	})

	Describe("Compile", func() {
		It("uploads the package archive to the blobstore and then compiles the package with the agent", func() {
			pkgDeps := map[string]PackageRef{
				"fake-package-name-dep": pkgDependencyRef,
			}
			packageRef, err := remotePackageCompiler.Compile(pkg, pkgDeps)
			Expect(err).ToNot(HaveOccurred())
			Expect(packageRef).To(Equal(PackageRef{
				Name:    "fake-package-name",
				Version: "fake-package-version",
				Archive: BlobRef{
					BlobstoreID: "fake-compiled-package-blob-id",
					SHA1:        "fake-compiled-package-sha1",
				},
			}))
		})
	})
})
