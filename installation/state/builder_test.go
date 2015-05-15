package state_test

import (
	. "github.com/cloudfoundry/bosh-init/installation/state"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_deployment_release "github.com/cloudfoundry/bosh-init/deployment/release/mocks"
	mock_state_job "github.com/cloudfoundry/bosh-init/state/job/mocks"
	mock_template "github.com/cloudfoundry/bosh-init/templatescompiler/mocks"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	biindex "github.com/cloudfoundry/bosh-init/index"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
	bistatejob "github.com/cloudfoundry/bosh-init/state/job"
	bitemplate "github.com/cloudfoundry/bosh-init/templatescompiler"
	fakeboshblob "github.com/cloudfoundry/bosh-utils/blobstore/fakes"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	fakeboshcmd "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeboshsys "github.com/cloudfoundry/bosh-utils/system/fakes"

	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("Builder", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		mockReleaseJobResolver *mock_deployment_release.MockJobResolver
		mockDependencyCompiler *mock_state_job.MockDependencyCompiler
		mockJobListRenderer    *mock_template.MockJobListRenderer
		fakeCompressor         *fakeboshcmd.FakeCompressor
		fakeBlobstore          *fakeboshblob.FakeBlobstore

		fakeFS        *fakeboshsys.FakeFileSystem
		templatesRepo bitemplate.TemplatesRepo

		logger boshlog.Logger

		builder Builder

		releaseJob bireljob.Job

		manifest  biinstallmanifest.Manifest
		fakeStage *fakebiui.FakeStage

		releasePackage1 *birelpkg.Package
		releasePackage2 *birelpkg.Package

		expectJobResolve *gomock.Call
		expectCompile    *gomock.Call
		expectJobRender  *gomock.Call
	)

	BeforeEach(func() {
		mockReleaseJobResolver = mock_deployment_release.NewMockJobResolver(mockCtrl)
		mockDependencyCompiler = mock_state_job.NewMockDependencyCompiler(mockCtrl)
		mockJobListRenderer = mock_template.NewMockJobListRenderer(mockCtrl)
		fakeCompressor = fakeboshcmd.NewFakeCompressor()
		fakeBlobstore = fakeboshblob.NewFakeBlobstore()

		fakeFS = fakeboshsys.NewFakeFileSystem()
		index := biindex.NewInMemoryIndex()
		templatesRepo = bitemplate.NewTemplatesRepo(index)

		logger = boshlog.NewLogger(boshlog.LevelNone)

		fakeStage = fakebiui.NewFakeStage()

		manifest = biinstallmanifest.Manifest{
			Name: "fake-installation-name",
			Template: biinstallmanifest.ReleaseJobRef{
				Name:    "fake-cpi-job-name",
				Release: "fake-cpi-release-name",
			},
			Properties: biproperty.Map{
				"fake-installation-property": "fake-installation-property-value",
			},
		}

		releasePackage1 = &birelpkg.Package{
			Name:          "fake-release-package-name-1",
			Fingerprint:   "fake-release-package-fingerprint-1",
			SHA1:          "fake-release-package-sha1-1",
			Dependencies:  []*birelpkg.Package{},
			ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name-1",
		}

		releasePackage2 = &birelpkg.Package{
			Name:          "fake-release-package-name-2",
			Fingerprint:   "fake-release-package-fingerprint-2",
			SHA1:          "fake-release-package-sha1-2",
			Dependencies:  []*birelpkg.Package{releasePackage1},
			ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name-2",
		}

		releaseJob = bireljob.Job{
			Name:          "cpi",
			Fingerprint:   "fake-release-job-fingerprint",
			SHA1:          "fake-release-job-sha1",
			ExtractedPath: "/extracted-release-path/extracted_jobs/cpi",
			Templates: map[string]string{
				"cpi.erb":     "bin/cpi",
				"cpi.yml.erb": "config/cpi.yml",
			},
			PackageNames: []string{releasePackage2.Name},
			Packages:     []*birelpkg.Package{releasePackage2},
			Properties:   map[string]bireljob.PropertyDefinition{},
		}
	})

	JustBeforeEach(func() {
		builder = NewBuilder(
			mockReleaseJobResolver,
			mockDependencyCompiler,
			mockJobListRenderer,
			fakeCompressor,
			fakeBlobstore,
			templatesRepo,
		)

		expectJobResolve = mockReleaseJobResolver.EXPECT().Resolve("fake-cpi-job-name", "fake-cpi-release-name").Return(releaseJob, nil).AnyTimes()

		releaseJobs := []bireljob.Job{releaseJob}
		compiledPackageRefs := []bistatejob.CompiledPackageRef{
			{
				Name:        "fake-release-package-name-1",
				Version:     "fake-release-package-fingerprint-1",
				BlobstoreID: "fake-compiled-package-blobstore-id-1",
				SHA1:        "fake-compiled-package-sha1-1",
			},
			{
				Name:        "fake-release-package-name-2",
				Version:     "fake-release-package-fingerprint-2",
				BlobstoreID: "fake-compiled-package-blobstore-id-2",
				SHA1:        "fake-compiled-package-sha1-2",
			},
		}
		expectCompile = mockDependencyCompiler.EXPECT().Compile(releaseJobs, fakeStage).Return(compiledPackageRefs, nil).AnyTimes()

		jobProperties := biproperty.Map{
			"fake-installation-property": "fake-installation-property-value",
		}
		globalProperties := biproperty.Map{}
		deploymentName := "fake-installation-name"

		renderedJobList := bitemplate.NewRenderedJobList()
		renderedJobList.Add(bitemplate.NewRenderedJob(releaseJob, "/fake-rendered-job-cpi", fakeFS, logger))

		expectJobRender = mockJobListRenderer.EXPECT().Render(releaseJobs, jobProperties, globalProperties, deploymentName).Return(renderedJobList, nil).AnyTimes()

		fakeCompressor.CompressFilesInDirTarballPath = "/fake-rendered-job-tarball-cpi.tgz"

		fakeBlobstore.CreateBlobIDs = []string{"fake-rendered-job-tarball-blobstore-id-cpi"}
		fakeBlobstore.CreateFingerprints = []string{"fake-rendered-job-tarball-sha1-cpi"}
	})

	Describe("Build", func() {
		It("finds the cpi job in the release specified by the installation manifest", func() {
			expectJobResolve.Times(1)

			_, err := builder.Build(manifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())
		})

		It("compiles the dependencies of the cpi job", func() {
			expectCompile.Times(1)

			_, err := builder.Build(manifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())
		})

		It("rendered the cpi job templates", func() {
			expectJobRender.Times(1)

			_, err := builder.Build(manifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())
		})

		It("logs compile & render stages", func() {
			_, err := builder.Build(manifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]fakebiui.PerformCall{
				// compile stages not produced by mockDependencyCompiler
				{Name: "Rendering job templates"},
			}))
		})

		It("compresses and uploads the rendered cpi job, deleting the local tarball afterward", func() {
			_, err := builder.Build(manifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCompressor.CompressFilesInDirDir).To(Equal("/fake-rendered-job-cpi"))
			Expect(fakeBlobstore.CreateFileNames).To(Equal([]string{"/fake-rendered-job-tarball-cpi.tgz"}))
			Expect(fakeCompressor.CleanUpTarballPath).To(Equal("/fake-rendered-job-tarball-cpi.tgz"))
		})

		It("returns a new installation state with the compiled packages and rendered jobs", func() {
			state, err := builder.Build(manifest, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(state.RenderedCPIJob()).To(Equal(biinstalljob.RenderedJobRef{
				Name:        "cpi",
				Version:     "fake-release-job-fingerprint",
				BlobstoreID: "fake-rendered-job-tarball-blobstore-id-cpi",
				SHA1:        "fake-rendered-job-tarball-sha1-cpi",
			}))

			Expect(state.CompiledPackages()).To(ContainElement(biinstallpkg.CompiledPackageRef{
				Name:        "fake-release-package-name-1",
				Version:     "fake-release-package-fingerprint-1",
				BlobstoreID: "fake-compiled-package-blobstore-id-1",
				SHA1:        "fake-compiled-package-sha1-1",
			}))
			Expect(state.CompiledPackages()).To(ContainElement(biinstallpkg.CompiledPackageRef{
				Name:        "fake-release-package-name-2",
				Version:     "fake-release-package-fingerprint-2",
				BlobstoreID: "fake-compiled-package-blobstore-id-2",
				SHA1:        "fake-compiled-package-sha1-2",
			}))
			Expect(state.CompiledPackages()).To(HaveLen(2))
		})

		Context("when the release does not contain a 'cpi' job", func() {
			JustBeforeEach(func() {
				expectJobResolve.Return(bireljob.Job{}, bosherr.Error("fake-job-resolve-error")).Times(1)
			})

			It("returns an error", func() {
				_, err := builder.Build(manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-job-resolve-error"))
			})
		})

		Context("when package compilation fails", func() {
			JustBeforeEach(func() {
				expectCompile.Return([]bistatejob.CompiledPackageRef{}, bosherr.Error("fake-compile-package-2-error")).Times(1)
			})

			It("returns an error", func() {
				_, err := builder.Build(manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-package-2-error"))
			})
		})
	})
})
