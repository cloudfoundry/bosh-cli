package cpi_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/cpi"

	"code.google.com/p/gomock/gomock"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"

	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
)

var _ = Describe("Deployment", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		manifest                  bmmanifest.CPIDeploymentManifest
		mockRegistryServerManager *mock_registry.MockServerManager
		mockRegistryServer        *mock_registry.MockServer
		fakeInstaller             *fakebmcpi.FakeInstaller

		deployment Deployment

		directorID = "fake-director-id"
	)
	BeforeEach(func() {
		manifest = bmmanifest.CPIDeploymentManifest{}

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)
		mockRegistryServer = mock_registry.NewMockServer(mockCtrl)

		fakeInstaller = fakebmcpi.NewFakeInstaller()

		deployment = NewDeployment(manifest, mockRegistryServerManager, fakeInstaller, directorID)
	})

	Describe("ExtractRelease & Install", func() {
		var (
			fakeCloud *fakebmcloud.FakeCloud
		)

		BeforeEach(func() {
			fakeCloud = fakebmcloud.NewFakeCloud()
		})

		Context("when ExtractRelease has been called", func() {
			var (
				releaseTarballPath = "fake-release-tarball-path"
				fakeRelease        *fakebmrel.FakeRelease
			)

			BeforeEach(func() {
				fakeRelease = fakebmrel.NewFakeRelease()

				fakeInstaller.SetExtractBehavior(releaseTarballPath, func(_ string) (bmrel.Release, error) {
					return fakeRelease, nil
				})

				release, err := deployment.ExtractRelease(releaseTarballPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(Equal(fakeRelease))
			})

			It("requires extract to be called first & delegates to CPIInstaller.Install", func() {
				fakeInstaller.SetInstallBehavior(manifest, fakeRelease, directorID, fakeCloud, nil)

				cloud, err := deployment.Install()
				Expect(err).ToNot(HaveOccurred())
				Expect(cloud).To(Equal(fakeCloud))

				Expect(fakeInstaller.InstallInputs).To(Equal([]fakebmcpi.InstallInput{
					{
						Deployment: manifest,
						Release:    fakeRelease,
						DirectorID: directorID,
					},
				}))
			})

			Context("when the release has already been deleted", func() {
				BeforeEach(func() {
					fakeRelease.Delete()
				})

				It("returns an error", func() {
					_, err := deployment.Install()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Extracted CPI release not found"))

					Expect(fakeInstaller.InstallInputs).To(HaveLen(0))
				})
			})
		})

		Context("when ExtractRelease has not been called", func() {
			It("returns an error", func() {
				_, err := deployment.Install()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("CPI release must be extracted before it can be installed"))

				Expect(fakeInstaller.InstallInputs).To(HaveLen(0))
			})
		})
	})

	Describe("StartJobs", func() {
		Context("when registry config is not empty", func() {
			BeforeEach(func() {
				manifest.Registry = bmmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
				deployment = NewDeployment(manifest, mockRegistryServerManager, fakeInstaller, directorID)
			})

			It("starts the registry", func() {
				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)

				err := deployment.StartJobs()
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when starting registry fails", func() {
				BeforeEach(func() {
					mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(nil, errors.New("fake-registry-start-error"))
				})

				It("returns an error", func() {
					err := deployment.StartJobs()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-registry-start-error"))
				})
			})
		})

		Context("when registry config is empty", func() {
			BeforeEach(func() {
				manifest.Registry = bmmanifest.Registry{}
				deployment = NewDeployment(manifest, mockRegistryServerManager, fakeInstaller, directorID)
			})

			It("does not start the registry", func() {
				err := deployment.StartJobs()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("StopJobs", func() {
		Context("when registry has been started", func() {
			BeforeEach(func() {
				manifest.Registry = bmmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
				deployment = NewDeployment(manifest, mockRegistryServerManager, fakeInstaller, directorID)

				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)
				err := deployment.StartJobs()
				Expect(err).ToNot(HaveOccurred())
			})

			It("stops the registry", func() {
				mockRegistryServer.EXPECT().Stop()

				err := deployment.StopJobs()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when registry is configured but has not been started", func() {
			BeforeEach(func() {
				manifest.Registry = bmmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
				deployment = NewDeployment(manifest, mockRegistryServerManager, fakeInstaller, directorID)
			})

			It("returns an error", func() {
				err := deployment.StopJobs()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("CPI jobs must be started before they can be stopped"))
			})
		})

		Context("when registry config is empty", func() {
			BeforeEach(func() {
				manifest.Registry = bmmanifest.Registry{}
				deployment = NewDeployment(manifest, mockRegistryServerManager, fakeInstaller, directorID)
			})

			It("does not stop the registry", func() {
				err := deployment.StopJobs()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
