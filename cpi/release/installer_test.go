package release_test

import (
	"code.google.com/p/gomock/gomock"
	"errors"
	"github.com/cloudfoundry/bosh-init/cpi/release"
	biinstallation "github.com/cloudfoundry/bosh-init/installation"
	biinstallationmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	"github.com/cloudfoundry/bosh-init/installation/mocks"
	"github.com/cloudfoundry/bosh-init/ui"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Installer", func() {
	Describe("WithInstalledCpiRelease", func() {
		var (
			mockCtrl                       *gomock.Controller
			mockInstaller                  *mocks.MockInstaller
			installationManifest           biinstallationmanifest.Manifest
			installStage                   *fakeui.FakeStage
			installation                   *mocks.MockInstallation
			expectedInstallPackagesAndJobs *gomock.Call
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		BeforeEach(func() {
			mockInstaller = mocks.NewMockInstaller(mockCtrl)
			installationManifest = biinstallationmanifest.Manifest{}
			installStage = fakeui.NewFakeStage()
			installation = mocks.NewMockInstallation(mockCtrl)

			expectedInstallPackagesAndJobs = mockInstaller.EXPECT().InstallPackagesAndJobs(installationManifest, gomock.Any())
		})

		It("should install the CPI and call the passed in function with the installation", func() {
			cpiInstaller := release.CpiInstaller{
				Installer: mockInstaller,
			}

			expectedInstallPackagesAndJobs.Return(installation, nil)

			var installationArgumentToFunction biinstallation.Installation
			err := cpiInstaller.WithInstalledCpiRelease(installationManifest, installStage, func(installation biinstallation.Installation) error {
				installationArgumentToFunction = installation
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(installationArgumentToFunction).ToNot(BeNil())
			Expect(installationArgumentToFunction).To(Equal(installation))

		})

		It("starts an 'installing CPI stage' and passes it to the installer", func() {
			cpiInstaller := release.CpiInstaller{
				Installer: mockInstaller,
			}

			var stageForInstallPackagesAndJobs ui.Stage
			expectedInstallPackagesAndJobs.Do(func(manifest biinstallationmanifest.Manifest, stage ui.Stage) (biinstallation.Installation, error) {
				stageForInstallPackagesAndJobs = stage
				return installation, nil
			})

			err := cpiInstaller.WithInstalledCpiRelease(installationManifest, installStage, func(installation biinstallation.Installation) error {
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(stageForInstallPackagesAndJobs).To(fakeui.BeASubstageOf(installStage))

			Expect(installStage.PerformCalls).To(Equal(
				[]*fakeui.PerformCall{
					{
						Name:  "installing CPI",
						Stage: fakeui.NewFakeStage(),
					},
				},
			))
		})

		Context("when installing the cpi fails", func() {
			It("returns the error", func() {
				cpiInstaller := release.CpiInstaller{
					Installer: mockInstaller,
				}

				expectedInstallPackagesAndJobs.Return(nil, errors.New("couldn't install that"))

				err := cpiInstaller.WithInstalledCpiRelease(installationManifest, installStage, func(installation biinstallation.Installation) error {
					return nil
				})

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
