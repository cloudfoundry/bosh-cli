package installation_test

import (
	. "github.com/cloudfoundry/bosh-init/installation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_install_job "github.com/cloudfoundry/bosh-init/installation/job/mocks"
	mock_install "github.com/cloudfoundry/bosh-init/installation/mocks"
	mock_install_pkg "github.com/cloudfoundry/bosh-init/installation/pkg/mocks"
	mock_registry "github.com/cloudfoundry/bosh-init/registry/mocks"

	biproperty "github.com/cloudfoundry/bosh-utils/property"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"errors"

	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"
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
		mockJobRenderer      *mock_install.MockJobRenderer
		mockJobResolver      *mock_install.MockJobResolver
		mockPackageCompiler  *mock_install.MockPackageCompiler
		mockPackageInstaller *mock_install_pkg.MockInstaller
		mockJobInstaller     *mock_install_job.MockInstaller

		mockRegistryServerManager *mock_registry.MockServerManager

		logger boshlog.Logger

		packagesPath string
		installer    Installer
		target       Target
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockJobRenderer = mock_install.NewMockJobRenderer(mockCtrl)
		mockJobResolver = mock_install.NewMockJobResolver(mockCtrl)
		mockPackageCompiler = mock_install.NewMockPackageCompiler(mockCtrl)
		mockPackageInstaller = mock_install_pkg.NewMockInstaller(mockCtrl)
		mockJobInstaller = mock_install_job.NewMockInstaller(mockCtrl)

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)

		packagesPath = "/path/to/installed/packages"
		target = NewTarget("fake-installation-path")
	})

	JustBeforeEach(func() {
		installer = NewInstaller(
			target,
			mockJobRenderer,
			mockJobResolver,
			mockPackageCompiler,
			packagesPath,
			mockPackageInstaller,
			mockJobInstaller,
			mockRegistryServerManager,
			logger,
		)
	})

	Describe("InstallPackagesAndJobs", func() {
		var (
			installationManifest biinstallmanifest.Manifest
			fakeStage            *fakebiui.FakeStage

			installedJob biinstalljob.InstalledJob

			expectedResolveJobsFrom     *gomock.Call
			expectedPackageCompilerFrom *gomock.Call
			expectedRenderAndUploadFrom *gomock.Call
			expectPackageInstall        *gomock.Call
			expectJobInstall            *gomock.Call
		)

		BeforeEach(func() {
			installationManifest = biinstallmanifest.Manifest{
				Name:       "fake-installation-name",
				Properties: biproperty.Map{},
			}

			fakeStage = fakebiui.NewFakeStage()

			installedJob = biinstalljob.NewInstalledJob(biinstalljob.RenderedJobRef{Name: "cpi"}, "/extracted-release-path/cpi")
		})

		JustBeforeEach(func() {
			renderedCPIJob := biinstalljob.RenderedJobRef{
				Name:        "cpi",
				Version:     "fake-release-job-fingerprint",
				BlobstoreID: "fake-rendered-job-blobstore-id",
				SHA1:        "fake-rendered-job-blobstore-id",
			}

			compiledPackageRef := biinstallpkg.CompiledPackageRef{
				Name:        "fake-release-package-name",
				Version:     "fake-release-package-fingerprint",
				BlobstoreID: "fake-compiled-package-blobstore-id",
				SHA1:        "fake-compiled-package-blobstore-id",
			}
			compiledPackages := []biinstallpkg.CompiledPackageRef{compiledPackageRef}

			releaseJobs := []bireljob.Job{}
			renderedJobRefs := []biinstalljob.RenderedJobRef{renderedCPIJob}
			expectedResolveJobsFrom = mockJobResolver.EXPECT().From(installationManifest).Return(releaseJobs, nil).AnyTimes()
			expectedPackageCompilerFrom = mockPackageCompiler.EXPECT().For(releaseJobs, packagesPath, fakeStage).Return(compiledPackages, nil).AnyTimes()
			expectedRenderAndUploadFrom = mockJobRenderer.EXPECT().RenderAndUploadFrom(installationManifest, releaseJobs, fakeStage).Return(renderedJobRefs, nil).AnyTimes()

			expectPackageInstall = mockPackageInstaller.EXPECT().Install(compiledPackageRef, packagesPath).AnyTimes()

			expectJobInstall = mockJobInstaller.EXPECT().Install(renderedCPIJob, fakeStage).Return(installedJob, nil).AnyTimes()
		})

		It("compiles and installs the jobs' packages", func() {
			expectPackageInstall.Times(1)

			_, err := installer.InstallPackagesAndJobs(installationManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
		})

		It("installs the rendered jobs", func() {
			expectJobInstall.Times(1)

			_, err := installer.InstallPackagesAndJobs(installationManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the installation", func() {
			installation, err := installer.InstallPackagesAndJobs(installationManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			expectedInstallation := NewInstallation(
				target,
				installedJob,
				installationManifest,
				mockRegistryServerManager,
			)

			Expect(installation).To(Equal(expectedInstallation))
		})
	})

	Describe("Cleanup", func() {
		It("cleans up installed jobs", func() {
			installation := mock_install.NewMockInstallation(mockCtrl)
			installationJob := biinstalljob.InstalledJob{}
			installation.EXPECT().Job().Return(installationJob)

			mockJobInstaller.EXPECT().Cleanup(installationJob)

			err := installer.Cleanup(installation)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns errors when cleaning up installed jobs", func() {
			installation := mock_install.NewMockInstallation(mockCtrl)
			installationJob := biinstalljob.InstalledJob{}
			installation.EXPECT().Job().Return(installationJob)

			mockJobInstaller.EXPECT().Cleanup(installationJob).Return(errors.New("nope"))

			err := installer.Cleanup(installation)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("nope"))
		})
	})
})
