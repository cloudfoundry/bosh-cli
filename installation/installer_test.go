package installation_test

import (
	"os"

	. "github.com/cloudfoundry/bosh-init/installation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_install_job "github.com/cloudfoundry/bosh-init/installation/job/mocks"
	mock_install_pkg "github.com/cloudfoundry/bosh-init/installation/pkg/mocks"
	mock_install_state "github.com/cloudfoundry/bosh-init/installation/state/mocks"
	mock_registry "github.com/cloudfoundry/bosh-init/registry/mocks"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	biinstallstate "github.com/cloudfoundry/bosh-init/installation/state"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
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
		fakeFS *fakesys.FakeFileSystem

		mockStateBuilder     *mock_install_state.MockBuilder
		mockPackageInstaller *mock_install_pkg.MockInstaller
		mockJobInstaller     *mock_install_job.MockInstaller

		mockRegistryServerManager *mock_registry.MockServerManager

		logger boshlog.Logger

		packagesPath           string
		deploymentManifestPath string
		installer              Installer
		target                 Target
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()

		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockStateBuilder = mock_install_state.NewMockBuilder(mockCtrl)
		mockPackageInstaller = mock_install_pkg.NewMockInstaller(mockCtrl)
		mockJobInstaller = mock_install_job.NewMockInstaller(mockCtrl)

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)

		packagesPath = "/path/to/installed/packages"
		deploymentManifestPath = "/path/to/manifest.yml"
		target = NewTarget("fake-installation-path")
	})

	JustBeforeEach(func() {
		installer = NewInstaller(
			target,
			fakeFS,
			mockStateBuilder,
			packagesPath,
			mockPackageInstaller,
			mockJobInstaller,
			mockRegistryServerManager,
			logger,
		)
	})

	Describe("Install", func() {
		var (
			installationManifest biinstallmanifest.Manifest
			fakeStage            *fakebiui.FakeStage

			installedJob biinstalljob.InstalledJob

			expectStateBuild     *gomock.Call
			expectPackageInstall *gomock.Call
			expectJobInstall     *gomock.Call
		)

		BeforeEach(func() {
			fakeFS.WriteFileString(deploymentManifestPath, "")

			installationManifest = biinstallmanifest.Manifest{
				Name:       "fake-installation-name",
				Properties: biproperty.Map{},
			}

			fakeStage = fakebiui.NewFakeStage()

			installedJob = biinstalljob.InstalledJob{
				Name: "cpi",
				Path: "/extracted-release-path/cpi",
			}
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

			state := biinstallstate.NewState(renderedCPIJob, compiledPackages)

			expectStateBuild = mockStateBuilder.EXPECT().Build(installationManifest, fakeStage).Return(state, nil).AnyTimes()

			expectPackageInstall = mockPackageInstaller.EXPECT().Install(compiledPackageRef, packagesPath).AnyTimes()

			expectJobInstall = mockJobInstaller.EXPECT().Install(renderedCPIJob, fakeStage).Return(installedJob, nil).AnyTimes()

			fakeFS.MkdirAll("/extracted-release-path", os.FileMode(0750))
		})

		It("builds a new installation state", func() {
			expectStateBuild.Times(1)

			_, err := installer.Install(installationManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
		})

		It("installs the compiled packages specified by the state", func() {
			expectPackageInstall.Times(1)

			_, err := installer.Install(installationManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
		})

		It("installs the rendered jobs specified by the state", func() {
			expectJobInstall.Times(1)

			_, err := installer.Install(installationManifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the installation", func() {
			installation, err := installer.Install(installationManifest, fakeStage)
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
})
