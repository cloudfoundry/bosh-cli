package state_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	biagentclient "github.com/cloudfoundry/bosh-agent/v2/agentclient"
	"github.com/cloudfoundry/bosh-cli/v7/agentclient/agentclientfakes"
	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state"
	biindex "github.com/cloudfoundry/bosh-cli/v7/index"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	bistatepkg "github.com/cloudfoundry/bosh-cli/v7/state/pkg"
)

var _ = Describe("RemotePackageCompiler", func() {
	var (
		packageRepo bistatepkg.CompiledPackageRepo

		mockBlobstore   *blobstorefakes.FakeBlobstore
		mockAgentClient *agentclientfakes.FakeAgentClient

		archivePath = "fake-archive-path"

		remotePackageCompiler bistatepkg.Compiler
	)

	BeforeEach(func() {
		mockBlobstore = &blobstorefakes.FakeBlobstore{}
		mockAgentClient = &agentclientfakes.FakeAgentClient{}

		index := biindex.NewInMemoryIndex()
		packageRepo = bistatepkg.NewCompiledPackageRepo(index)
		remotePackageCompiler = NewRemotePackageCompiler(mockBlobstore, mockAgentClient, packageRepo)
	})

	Describe("Compile", func() {
		Context("when package is not compiled", func() {
			var (
				pkgDependency *boshpkg.Package
				pkg           *boshpkg.Package

				compiledPackages map[bistatepkg.CompiledPackageRecord]*boshpkg.Package
			)

			BeforeEach(func() {
				pkgDependency = boshpkg.NewPackage(NewResource(
					"fake-package-name-dep", "fake-package-fingerprint-dep", nil), nil)

				pkg = boshpkg.NewPackage(NewResourceWithBuiltArchive(
					"fake-package-name", "fake-package-fingerprint", archivePath, "fake-source-package-sha1"), []string{"fake-package-name-dep"})
				err := pkg.AttachDependencies([]*boshpkg.Package{pkgDependency})
				Expect(err).ToNot(HaveOccurred())

				depRecord1 := bistatepkg.CompiledPackageRecord{
					BlobID:   "fake-compiled-package-blob-id-dep",
					BlobSHA1: "fake-compiled-package-sha1-dep",
				}

				compiledPackages = map[bistatepkg.CompiledPackageRecord]*boshpkg.Package{
					depRecord1: pkgDependency,
				}
			})

			JustBeforeEach(func() {
				// add compiled packages to the repo
				for record, dependency := range compiledPackages {
					err := packageRepo.Save(dependency, record)
					Expect(err).ToNot(HaveOccurred())
				}

				compiledPackageRef := biagentclient.BlobRef{
					Name:        "fake-package-name",
					Version:     "fake-package-version",
					BlobstoreID: "fake-compiled-package-blob-id",
					SHA1:        "fake-compiled-package-sha1",
				}

				mockBlobstore.AddReturns("fake-source-package-blob-id", nil)
				mockAgentClient.CompilePackageReturns(compiledPackageRef, nil)
			})

			It("uploads the package archive to the blobstore and then compiles the package with the agent", func() {
				var order []string
				mockBlobstore.AddStub = func(path string) (string, error) {
					order = append(order, "Add")
					return "fake-source-package-blob-id", nil
				}
				mockAgentClient.CompilePackageStub = func(source biagentclient.BlobRef, deps []biagentclient.BlobRef) (biagentclient.BlobRef, error) {
					order = append(order, "CompilePackage")
					return biagentclient.BlobRef{
						Name:        "fake-package-name",
						Version:     "fake-package-version",
						BlobstoreID: "fake-compiled-package-blob-id",
						SHA1:        "fake-compiled-package-sha1",
					}, nil
				}

				compiledPackageRecord, _, err := remotePackageCompiler.Compile(pkg)
				Expect(order).To(Equal([]string{"Add", "CompilePackage"}))
				Expect(err).ToNot(HaveOccurred())
				Expect(compiledPackageRecord).To(Equal(bistatepkg.CompiledPackageRecord{
					BlobID:   "fake-compiled-package-blob-id",
					BlobSHA1: "fake-compiled-package-sha1",
				}))
			})

			It("saves the compiled package ref in the package repo", func() {
				compiledPackageRecord, _, err := remotePackageCompiler.Compile(pkg)
				Expect(err).ToNot(HaveOccurred())

				record, found, err := packageRepo.Find(pkg)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(record).To(Equal(compiledPackageRecord))
			})

			Context("when the dependencies are not in the repo", func() {
				BeforeEach(func() {
					compiledPackages = map[bistatepkg.CompiledPackageRecord]*boshpkg.Package{}
				})

				It("returns an error", func() {
					_, _, err := remotePackageCompiler.Compile(pkg)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Remote compilation failure: Package 'fake-package-name/fake-package-fingerprint' requires package 'fake-package-name-dep/fake-package-fingerprint-dep', but it has not been compiled"))
				})
			})
		})

		Context("when package is compiled", func() {
			var (
				pkgDependency *boshpkg.CompiledPackage
				pkg           *boshpkg.CompiledPackage
			)

			BeforeEach(func() {
				pkgDependency = boshpkg.NewCompiledPackageWithoutArchive(
					"fake-package-name-dep", "fake-package-fingerprint-dep", "", "", nil)

				pkg = boshpkg.NewCompiledPackageWithArchive(
					"fake-package-name", "fake-package-fingerprint", "", archivePath, "fake-source-package-sha1", []string{"fake-package-name-dep"})
				err := pkg.AttachDependencies([]*boshpkg.CompiledPackage{pkgDependency})
				Expect(err).ToNot(HaveOccurred())
			})

			It("should skip compilation but still add blobstore package", func() {
				err := packageRepo.Save(pkgDependency, bistatepkg.CompiledPackageRecord{
					BlobID:   "fake-compiled-package-blob-id-dep",
					BlobSHA1: "fake-compiled-package-sha1-dep",
				})
				Expect(err).ToNot(HaveOccurred())

				mockBlobstore.AddReturns("fake-source-package-blob-id", nil)

				compiledPackageRecord, isAlreadyCompiled, err := remotePackageCompiler.Compile(pkg)
				Expect(err).ToNot(HaveOccurred())
				Expect(isAlreadyCompiled).To(Equal(true))
				Expect(compiledPackageRecord).To(Equal(bistatepkg.CompiledPackageRecord{
					BlobID:   "fake-source-package-blob-id",
					BlobSHA1: "fake-source-package-sha1",
				}))

				Expect(mockAgentClient.CompilePackageCallCount()).To(Equal(0))
			})
		})
	})
})
