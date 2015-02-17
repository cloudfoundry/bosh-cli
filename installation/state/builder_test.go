package state_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/state"

	"code.google.com/p/gomock/gomock"
	mock_deployment_release "github.com/cloudfoundry/bosh-micro-cli/deployment/release/mocks"
	mock_state_package "github.com/cloudfoundry/bosh-micro-cli/state/pkg/mocks"
	mock_template "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/mocks"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakeboshblob "github.com/cloudfoundry/bosh-agent/blobstore/fakes"
	fakeboshcmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakeboshsys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"

	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
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
		mockPackageCompiler    *mock_state_package.MockCompiler
		mockJobListRenderer    *mock_template.MockJobListRenderer
		fakeCompressor         *fakeboshcmd.FakeCompressor
		fakeBlobstore          *fakeboshblob.FakeBlobstore

		fakeFS        *fakeboshsys.FakeFileSystem
		templatesRepo bmtemplate.TemplatesRepo

		logger boshlog.Logger

		builder Builder

		releaseJob bmreljob.Job

		manifest  bminstallmanifest.Manifest
		fakeStage *fakebmui.FakeStage

		releasePackage1 *bmrelpkg.Package
		releasePackage2 *bmrelpkg.Package

		expectJobResolve  *gomock.Call
		expectCompilePkg1 *gomock.Call
		expectCompilePkg2 *gomock.Call
		expectJobRender   *gomock.Call
	)

	BeforeEach(func() {
		mockReleaseJobResolver = mock_deployment_release.NewMockJobResolver(mockCtrl)
		mockPackageCompiler = mock_state_package.NewMockCompiler(mockCtrl)
		mockJobListRenderer = mock_template.NewMockJobListRenderer(mockCtrl)
		fakeCompressor = fakeboshcmd.NewFakeCompressor()
		fakeBlobstore = fakeboshblob.NewFakeBlobstore()

		fakeFS = fakeboshsys.NewFakeFileSystem()
		index := bmindex.NewInMemoryIndex()
		templatesRepo = bmtemplate.NewTemplatesRepo(index)

		logger = boshlog.NewLogger(boshlog.LevelNone)

		fakeStage = fakebmui.NewFakeStage()

		manifest = bminstallmanifest.Manifest{
			Name: "fake-installation-name",
			Template: bminstallmanifest.ReleaseJobRef{
				Name:    "fake-cpi-job-name",
				Release: "fake-cpi-release-name",
			},
			Properties: bmproperty.Map{
				"fake-installation-property": "fake-installation-property-value",
			},
		}

		releasePackage1 = &bmrelpkg.Package{
			Name:          "fake-release-package-name-1",
			Fingerprint:   "fake-release-package-fingerprint-1",
			SHA1:          "fake-release-package-sha1-1",
			Dependencies:  []*bmrelpkg.Package{},
			ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name-1",
		}

		releasePackage2 = &bmrelpkg.Package{
			Name:          "fake-release-package-name-2",
			Fingerprint:   "fake-release-package-fingerprint-2",
			SHA1:          "fake-release-package-sha1-2",
			Dependencies:  []*bmrelpkg.Package{releasePackage1},
			ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name-2",
		}

		releaseJob = bmreljob.Job{
			Name:          "cpi",
			Fingerprint:   "fake-release-job-fingerprint",
			SHA1:          "fake-release-job-sha1",
			ExtractedPath: "/extracted-release-path/extracted_jobs/cpi",
			Templates: map[string]string{
				"cpi.erb":     "bin/cpi",
				"cpi.yml.erb": "config/cpi.yml",
			},
			PackageNames: []string{releasePackage2.Name},
			Packages:     []*bmrelpkg.Package{releasePackage2},
			Properties:   map[string]bmreljob.PropertyDefinition{},
		}
	})

	JustBeforeEach(func() {
		builder = NewBuilder(
			mockReleaseJobResolver,
			mockPackageCompiler,
			mockJobListRenderer,
			fakeCompressor,
			fakeBlobstore,
			templatesRepo,
		)

		expectJobResolve = mockReleaseJobResolver.EXPECT().Resolve("fake-cpi-job-name", "fake-cpi-release-name").Return(releaseJob, nil).AnyTimes()

		compiledPackageRecord1 := bmstatepkg.CompiledPackageRecord{
			BlobID:   "fake-compiled-package-blobstore-id-1",
			BlobSHA1: "fake-compiled-package-sha1-1",
		}
		expectCompilePkg1 = mockPackageCompiler.EXPECT().Compile(releasePackage1).Return(compiledPackageRecord1, nil).AnyTimes()

		compiledPackageRecord2 := bmstatepkg.CompiledPackageRecord{
			BlobID:   "fake-compiled-package-blobstore-id-2",
			BlobSHA1: "fake-compiled-package-sha1-2",
		}
		expectCompilePkg2 = mockPackageCompiler.EXPECT().Compile(releasePackage2).Return(compiledPackageRecord2, nil).AnyTimes()

		releaseJobs := []bmreljob.Job{releaseJob}
		jobProperties := bmproperty.Map{
			"fake-installation-property": "fake-installation-property-value",
		}
		globalProperties := bmproperty.Map{}
		deploymentName := "fake-installation-name"

		renderedJobList := bmtemplate.NewRenderedJobList()
		renderedJobList.Add(bmtemplate.NewRenderedJob(releaseJob, "/fake-rendered-job-cpi", fakeFS, logger))

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

		It("compiles the transitive package dependencies of the cpi job, in compilation order", func() {
			gomock.InOrder(
				expectCompilePkg1.Times(1),
				expectCompilePkg2.Times(1),
			)

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

			Expect(fakeStage.PerformCalls).To(Equal([]fakebmui.PerformCall{
				{
					Name: "Compiling package 'fake-release-package-name-1/fake-release-package-fingerprint-1'",
				},
				{
					Name: "Compiling package 'fake-release-package-name-2/fake-release-package-fingerprint-2'",
				},
				{
					Name: "Rendering job templates",
				},
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

			Expect(state.RenderedCPIJob()).To(Equal(bminstalljob.RenderedJobRef{
				Name:        "cpi",
				Version:     "fake-release-job-fingerprint",
				BlobstoreID: "fake-rendered-job-tarball-blobstore-id-cpi",
				SHA1:        "fake-rendered-job-tarball-sha1-cpi",
			}))

			Expect(state.CompiledPackages()).To(ContainElement(bminstallpkg.CompiledPackageRef{
				Name:        "fake-release-package-name-1",
				Version:     "fake-release-package-fingerprint-1",
				BlobstoreID: "fake-compiled-package-blobstore-id-1",
				SHA1:        "fake-compiled-package-sha1-1",
			}))
			Expect(state.CompiledPackages()).To(ContainElement(bminstallpkg.CompiledPackageRef{
				Name:        "fake-release-package-name-2",
				Version:     "fake-release-package-fingerprint-2",
				BlobstoreID: "fake-compiled-package-blobstore-id-2",
				SHA1:        "fake-compiled-package-sha1-2",
			}))
			Expect(state.CompiledPackages()).To(HaveLen(2))
		})

		Context("when the release does not contain a 'cpi' job", func() {
			JustBeforeEach(func() {
				expectJobResolve.Return(bmreljob.Job{}, bosherr.Error("fake-job-resolve-error")).Times(1)
			})

			It("returns an error", func() {
				_, err := builder.Build(manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-job-resolve-error"))
			})
		})

		Context("when package compilation fails", func() {
			JustBeforeEach(func() {
				expectCompilePkg2.Return(bmstatepkg.CompiledPackageRecord{}, bosherr.Error("fake-compile-package-2-error")).Times(1)
			})

			It("returns an error", func() {
				_, err := builder.Build(manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-package-2-error"))
			})
		})
	})
})
