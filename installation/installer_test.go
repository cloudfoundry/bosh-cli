package installation_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/installation"
	"github.com/cloudfoundry/bosh-cli/v7/installation/blobextract/blobextractfakes"
	"github.com/cloudfoundry/bosh-cli/v7/installation/installationfakes"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("Installer", func() {
	var (
		installationManifest biinstallmanifest.Manifest
		mockJobRenderer      *installationfakes.FakeJobRenderer
		mockJobResolver      *installationfakes.FakeJobResolver
		mockPackageCompiler  *installationfakes.FakePackageCompiler
		fakeExtractor        *blobextractfakes.FakeExtractor

		logger boshlog.Logger

		installer     Installer
		target        Target
		installedJobs []InstalledJob
	)

	BeforeEach(func() {
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, GinkgoWriter)

		mockJobRenderer = &installationfakes.FakeJobRenderer{}
		mockJobResolver = &installationfakes.FakeJobResolver{}
		mockPackageCompiler = &installationfakes.FakePackageCompiler{}
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
			fakeStage *testui.Stage

			renderedJobRefs []RenderedJobRef
			releaseJobs     []bireljob.Job
		)

		BeforeEach(func() {
			fakeStage = &testui.Stage{}
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
			mockJobResolver.FromReturns(releaseJobs, nil)
			mockPackageCompiler.ForReturns(compiledPackages, nil)
		})

		Context("success", func() {
			JustBeforeEach(func() {
				mockJobRenderer.RenderAndUploadFromReturns(renderedJobRefs, nil)
			})

			It("compiles and installs the jobs' packages", func() {
				_, err := installer.Install(installationManifest, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(mockJobResolver.FromCallCount()).To(Equal(1))
				Expect(mockJobResolver.FromArgsForCall(0)).To(Equal(installationManifest))

				Expect(mockPackageCompiler.ForCallCount()).To(Equal(1))
				actualJobs, actualStage := mockPackageCompiler.ForArgsForCall(0)
				Expect(actualJobs).To(Equal(releaseJobs))
				Expect(actualStage).To(Equal(fakeStage))

				Expect(mockJobRenderer.RenderAndUploadFromCallCount()).To(Equal(1))
				actualManifest, actualRenderJobs, actualRenderStage := mockJobRenderer.RenderAndUploadFromArgsForCall(0)
				Expect(actualManifest).To(Equal(installationManifest))
				Expect(actualRenderJobs).To(Equal(releaseJobs))
				Expect(actualRenderStage).To(Equal(fakeStage))
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
				mockJobRenderer.RenderAndUploadFromReturns([]RenderedJobRef{}, err)
			})
			It("should return an error", func() {
				_, err := installer.Install(installationManifest, fakeStage)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Install path traversal protection", func() {
		var fakeStage *testui.Stage

		BeforeEach(func() {
			fakeStage = &testui.Stage{}
		})

		It("returns error when a compiled package name contains path traversal", func() {
			maliciousRef := CompiledPackageRef{
				Name:        "../../path",
				Version:     "v1",
				BlobstoreID: "bid",
				SHA1:        "sha",
			}
			releaseJobs := []bireljob.Job{}
			mockJobResolver.FromReturns(releaseJobs, nil)
			mockPackageCompiler.ForReturns([]CompiledPackageRef{maliciousRef}, nil)

			_, err := installer.Install(installationManifest, fakeStage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("safe local path"))
		})

		It("returns error when a rendered job name contains path traversal", func() {
			releaseJobs := []bireljob.Job{}
			compiledPackages := []CompiledPackageRef{}
			jobRef := NewRenderedJobRef("../../path", "fp", "blob-id", "sha")
			mockJobResolver.FromReturns(releaseJobs, nil)
			mockPackageCompiler.ForReturns(compiledPackages, nil)
			mockJobRenderer.RenderAndUploadFromReturns([]RenderedJobRef{jobRef}, nil)

			_, err := installer.Install(installationManifest, fakeStage)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("safe local path"))
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
