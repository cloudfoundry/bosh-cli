package installation_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/installation"
	"github.com/cloudfoundry/bosh-cli/v7/installation/blobextract/blobextractfakes"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	mock_install "github.com/cloudfoundry/bosh-cli/v7/installation/mocks"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("Installer", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		installationManifest biinstallmanifest.Manifest
		mockJobRenderer      *mock_install.MockJobRenderer
		mockJobResolver      *mock_install.MockJobResolver
		mockPackageCompiler  *mock_install.MockPackageCompiler
		fakeExtractor        *blobextractfakes.FakeExtractor

		logger boshlog.Logger

		installer     Installer
		target        Target
		installedJobs []InstalledJob
	)

	BeforeEach(func() {
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, GinkgoWriter)

		mockJobRenderer = mock_install.NewMockJobRenderer(mockCtrl)
		mockJobResolver = mock_install.NewMockJobResolver(mockCtrl)
		mockPackageCompiler = mock_install.NewMockPackageCompiler(mockCtrl)
		fakeExtractor = &blobextractfakes.FakeExtractor{}

		target = NewTarget("fake-installation-path", "")
		installationManifest = biinstallmanifest.Manifest{
			Name:       "fake-installation-name",
			Properties: biproperty.Map{},
		}
		renderedCPIJob := NewRenderedJobRef("cpi", "fake-release-job-fingerprint", "fake-rendered-job-blobstore-id", "fake-rendered-job-blobstore-id")
		renderedCPIPluginJob := NewRenderedJobRef("cpi-plugin", "fake-release-job-fingerprint", "fake-rendered-job-blobstore-id", "fake-rendered-job-blobstore-id")

		installedJobs = make([]InstalledJob, 0)
		installedJobs = append(installedJobs, NewInstalledJob(renderedCPIJob, "/extracted-release-path/cpi"))
		installedJobs = append(installedJobs, NewInstalledJob(renderedCPIPluginJob,
			"/extracted-release-path/cpi-plugin"))
	})

	JustBeforeEach(func() {
		installer = NewInstaller(
			target,
			mockJobRenderer,
			mockJobResolver,
			mockPackageCompiler,
			fakeExtractor,
			logger,
		)
	})

	Describe("Install", func() {
		var (
			fakeStage *fakebiui.FakeStage

			renderedJobRefs []RenderedJobRef
			releaseJobs     []bireljob.Job
		)

		BeforeEach(func() {
			fakeStage = fakebiui.NewFakeStage()
		})

		JustBeforeEach(func() {
			ref := CompiledPackageRef{
				Name:        "fake-release-package-name",
				Version:     "fake-release-package-fingerprint",
				BlobstoreID: "fake-compiled-package-blobstore-id",
				SHA1:        "fake-compiled-package-blobstore-id",
			}
			compiledPackages := []CompiledPackageRef{ref}

			releaseJobs = []bireljob.Job{}

			renderedJobRefs = make([]RenderedJobRef, 0)
			for _, installedJob := range installedJobs {
				renderedJobRefs = append(renderedJobRefs, installedJob.RenderedJobRef)
			}
			mockJobResolver.EXPECT().From(installationManifest).Return(releaseJobs, nil).AnyTimes()
			mockPackageCompiler.EXPECT().For(releaseJobs, fakeStage).Return(compiledPackages, nil).AnyTimes()
		})

		Context("success", func() {
			JustBeforeEach(func() {
				mockJobRenderer.EXPECT().RenderAndUploadFrom(installationManifest, releaseJobs, fakeStage).Return(renderedJobRefs, nil).AnyTimes()
			})

			It("compiles and installs the jobs' packages", func() {
				_, err := installer.Install(installationManifest, fakeStage)
				Expect(err).NotTo(HaveOccurred())
			})

			It("installs the rendered jobs", func() {
				_, err := installer.Install(installationManifest, fakeStage)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the installation", func() {
				installation, err := installer.Install(installationManifest, fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(installation.Target().JobsPath()).To(Equal(target.JobsPath()))
			})
		})

		Context("when rendering jobs errors", func() {
			JustBeforeEach(func() {
				err := errors.New("OMG - no ruby found!!")
				mockJobRenderer.EXPECT().RenderAndUploadFrom(installationManifest, releaseJobs, fakeStage).Return([]RenderedJobRef{}, err).AnyTimes()
			})
			It("should return an error", func() {
				_, err := installer.Install(installationManifest, fakeStage)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Cleanup", func() {
		var installation Installation

		BeforeEach(func() {
			installation = NewInstallation(
				target,
				installedJobs,
				installationManifest,
			)
		})

		It("cleans up installed jobs", func() {
			err := installer.Cleanup(installation)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeExtractor.CleanupCallCount()).To(Equal(2))

			for i, installedJob := range installedJobs {
				blobstoreID, extractedBlobPath := fakeExtractor.CleanupArgsForCall(i)
				Expect(blobstoreID).To(Equal(installedJob.BlobstoreID))
				Expect(extractedBlobPath).To(Equal(installedJob.Path))
			}
		})

		It("returns errors when cleaning up installed jobs", func() {
			fakeExtractor.CleanupReturns(errors.New("nope"))

			err := installer.Cleanup(installation)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("nope"))
		})
	})
})
