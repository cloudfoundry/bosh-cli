package release_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	biinstallation "github.com/cloudfoundry/bosh-cli/v7/installation"
	"github.com/cloudfoundry/bosh-cli/v7/installation/installationfakes"
	biinstallationmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("Installer", func() {
	Describe("WithInstalledCpiRelease", func() {
		var (
			mockInstaller        *installationfakes.FakeInstaller
			mockInstallerFactory *installationfakes.FakeInstallerFactory
			installationManifest biinstallationmanifest.Manifest
			installStage         *testui.Stage
			installation         *installationfakes.FakeInstallation
			target               biinstallation.Target
		)

		BeforeEach(func() {
			mockInstaller = &installationfakes.FakeInstaller{}
			mockInstallerFactory = &installationfakes.FakeInstallerFactory{}

			installationManifest = biinstallationmanifest.Manifest{}
			installStage = &testui.Stage{}
			installation = &installationfakes.FakeInstallation{}

			target = biinstallation.NewTarget("fake-installation-path", "")
			mockInstallerFactory.NewInstallerReturns(mockInstaller)
			mockInstaller.CleanupReturns(nil)
		})

		It("should validate CPI release that include CPI and plugin releases", Pending, func() {})

		It("should install the CPI and call the passed in function with the installation", func() {
			cpiInstaller := release.CpiInstaller{
				InstallerFactory: mockInstallerFactory,
			}

			mockInstaller.InstallReturns(installation, nil)

			var installationArgumentToFunction biinstallation.Installation
			err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
				installationArgumentToFunction = installation
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(installationArgumentToFunction).ToNot(BeNil())
			Expect(installationArgumentToFunction).To(Equal(installation))

		})

		It("starts an 'installing CPI stage' and passes it to the installer", func() {
			cpiInstaller := release.CpiInstaller{
				InstallerFactory: mockInstallerFactory,
			}

			mockInstaller.InstallReturns(installation, nil)

			err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			_, stageForInstall := mockInstaller.InstallArgsForCall(0)
			Expect(stageForInstall).To(testui.BeASubstageOf(installStage))

			Expect(installStage.PerformCalls).To(ContainElement(
				&testui.PerformCall{
					Name:  "installing CPI",
					Stage: &testui.Stage{},
				},
			))
		})

		It("cleans up the installation afterwards", func() {
			cpiInstaller := release.CpiInstaller{
				InstallerFactory: mockInstallerFactory,
			}

			mockInstaller.InstallReturns(installation, nil)

			err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
				Expect(mockInstaller.CleanupCallCount()).To(Equal(0))
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(mockInstaller.CleanupCallCount()).To(Equal(1))
		})

		It("creates a stage for the cleanup", func() {
			cpiInstaller := release.CpiInstaller{
				InstallerFactory: mockInstallerFactory,
			}
			mockInstaller.InstallReturns(installation, nil)

			err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(installStage.PerformCalls).To(ContainElement(
				&testui.PerformCall{
					Name: "Cleaning up rendered CPI jobs",
				},
			))

		})

		Context("when installing the cpi fails", func() {
			It("returns the error", func() {
				cpiInstaller := release.CpiInstaller{
					InstallerFactory: mockInstallerFactory,
				}

				mockInstaller.InstallReturns(nil, errors.New("couldn't install that"))

				err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
					return nil
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("couldn't install that"))
				Expect(mockInstaller.CleanupCallCount()).To(Equal(0))
			})
		})

		Context("when the passed in function returns an error", func() {
			It("returns the error", func() {
				cpiInstaller := release.CpiInstaller{
					InstallerFactory: mockInstallerFactory,
				}

				mockInstaller.InstallReturns(installation, nil)

				err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
					return errors.New("My passed in function failed")
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("My passed in function failed"))
			})

			It("cleans up the installation", func() {
				cpiInstaller := release.CpiInstaller{
					InstallerFactory: mockInstallerFactory,
				}

				mockInstaller.InstallReturns(installation, nil)

				err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
					return errors.New("My passed in function failed")
				})
				Expect(err).To(HaveOccurred())
				Expect(mockInstaller.CleanupCallCount()).To(Equal(1))
			})
		})

		Context("when cleanup fails", func() {
			It("returns the error", func() {
				cpiInstaller := release.CpiInstaller{
					InstallerFactory: mockInstallerFactory,
				}

				mockInstaller.InstallReturns(installation, nil)
				mockInstaller.CleanupReturns(errors.New("cleanup failed"))

				err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
					return nil
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cleanup failed"))
			})

			It("returns the error from the function instead, if present", func() {
				cpiInstaller := release.CpiInstaller{
					InstallerFactory: mockInstallerFactory,
				}

				mockInstaller.InstallReturns(installation, nil)
				mockInstaller.CleanupReturns(errors.New("cleanup failed"))

				err := cpiInstaller.WithInstalledCpiRelease(installationManifest, target, installStage, func(installation biinstallation.Installation) error {
					return errors.New("My passed in function failed")
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("My passed in function failed"))
			})
		})
	})
})
