package installation_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/installation"

	"code.google.com/p/gomock/gomock"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"

	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

var _ = Describe("Installation", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		manifest                  bminstallmanifest.Manifest
		mockRegistryServerManager *mock_registry.MockServerManager
		mockRegistryServer        *mock_registry.MockServer

		target       Target
		installedJob bminstalljob.InstalledJob
	)

	var newInstalation = func() Installation {
		return NewInstallation(target, installedJob, manifest, mockRegistryServerManager)
	}

	BeforeEach(func() {
		manifest = bminstallmanifest.Manifest{}

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)
		mockRegistryServer = mock_registry.NewMockServer(mockCtrl)

		target = NewTarget("fake-installation-path")

		installedJob = bminstalljob.InstalledJob{
			Name: "cpi",
			Path: "fake-job-path",
		}
	})

	Describe("StartRegistry", func() {
		Context("when registry config is not empty", func() {
			BeforeEach(func() {
				manifest.Registry = bminstallmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
			})

			It("starts the registry", func() {
				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)

				err := newInstalation().StartRegistry()
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when starting registry fails", func() {
				BeforeEach(func() {
					mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(nil, errors.New("fake-registry-start-error"))
				})

				It("returns an error", func() {
					err := newInstalation().StartRegistry()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-registry-start-error"))
				})
			})
		})

		Context("when registry config is empty", func() {
			BeforeEach(func() {
				manifest.Registry = bminstallmanifest.Registry{}
			})

			It("does not start the registry", func() {
				err := newInstalation().StartRegistry()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("StopRegistry", func() {
		Context("when registry has been started", func() {
			var installation Installation

			BeforeEach(func() {
				manifest.Registry = bminstallmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}

				installation = newInstalation()

				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)
				err := installation.StartRegistry()
				Expect(err).ToNot(HaveOccurred())
			})

			It("stops the registry", func() {
				mockRegistryServer.EXPECT().Stop()

				err := installation.StopRegistry()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when registry is configured but has not been started", func() {
			BeforeEach(func() {
				manifest.Registry = bminstallmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
			})

			It("returns an error", func() {
				err := newInstalation().StopRegistry()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Registry must be started before it can be stopped"))
			})
		})

		Context("when registry config is empty", func() {
			BeforeEach(func() {
				manifest.Registry = bminstallmanifest.Registry{}
			})

			It("does not stop the registry", func() {
				err := newInstalation().StopRegistry()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
