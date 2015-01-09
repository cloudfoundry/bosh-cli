package installation_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/installation"

	"code.google.com/p/gomock/gomock"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"
	mock_release "github.com/cloudfoundry/bosh-micro-cli/release/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/installation/pkg/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
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
		fakeFS              *fakesys.FakeFileSystem
		fakeExtractor       *testfakes.FakeMultiResponseExtractor
		fakeReleaseCompiler *fakebmcomp.FakeReleaseCompiler
		fakeJobInstaller    *fakebminstalljob.FakeInstaller
		fakeUI              *fakebmui.FakeUI

		mockReleaseManager        *mock_release.MockManager
		mockRegistryServerManager *mock_registry.MockServerManager

		logger boshlog.Logger

		deploymentManifestPath string
		installer              Installer
		target                 Target
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseCompiler = fakebmcomp.NewFakeReleaseCompiler()
		fakeJobInstaller = fakebminstalljob.NewFakeInstaller()
		fakeUI = &fakebmui.FakeUI{}

		logger = boshlog.NewLogger(boshlog.LevelNone)

		mockReleaseManager = mock_release.NewMockManager(mockCtrl)
		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)

		deploymentManifestPath = "/path/to/manifest.yml"
		target = NewTarget("fake-installation-path")
	})

	JustBeforeEach(func() {
		installer = NewInstaller(
			target,
			fakeUI,
			mockReleaseManager,
			fakeReleaseCompiler,
			fakeJobInstaller,
			mockRegistryServerManager,
			logger,
		)
	})

	Describe("Install", func() {
		var (
			installationManifest bminstallmanifest.Manifest
			release              bmrel.Release
			releaseJob           bmrel.Job

			installedJob bminstalljob.InstalledJob

			expectFindByName *gomock.Call
		)

		BeforeEach(func() {
			fakeFS.WriteFileString(deploymentManifestPath, "")

			installationManifest = bminstallmanifest.Manifest{
				Name:          "fake-installation-name",
				Release:       "fake-release-name",
				RawProperties: map[interface{}]interface{}{},
			}

			releasePackage := &bmrel.Package{
				Name:          "fake-release-package-name",
				Fingerprint:   "fake-release-package-fingerprint",
				SHA1:          "fake-release-package-sha1",
				Dependencies:  []*bmrel.Package{},
				ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name",
			}

			releaseJob = bmrel.Job{
				Name:          "cpi",
				Fingerprint:   "fake-release-job-fingerprint",
				SHA1:          "fake-release-job-sha1",
				ExtractedPath: "/extracted-release-path/extracted_jobs/cpi",
				Templates: map[string]string{
					"cpi.erb":     "bin/cpi",
					"cpi.yml.erb": "config/cpi.yml",
				},
				PackageNames: []string{releasePackage.Name},
				Packages:     []*bmrel.Package{releasePackage},
				Properties:   map[string]bmrel.PropertyDefinition{},
			}

			installedJob = bminstalljob.InstalledJob{
				Name: "cpi",
				Path: "/extracted-release-path/cpi",
			}
		})

		JustBeforeEach(func() {
			releaseJobs := []bmrel.Job{releaseJob}
			releasePackages := append([]*bmrel.Package(nil), releaseJob.Packages...)
			release = bmrel.NewRelease(
				"fake-release-name",
				"fake-release-version",
				releaseJobs,
				releasePackages,
				"/extracted-release-path",
				fakeFS,
			)

			fakeJobInstaller.SetInstallBehavior(releaseJob, func(_ bmrel.Job) (bminstalljob.InstalledJob, error) {
				return installedJob, nil
			})

			fakeReleaseCompiler.SetCompileBehavior(release, installationManifest, nil)

			fakeFS.MkdirAll("/extracted-release-path", os.FileMode(0750))

			expectFindByName = mockReleaseManager.EXPECT().FindByName("fake-release-name").Return(release, true)
		})

		It("compiles the release", func() {
			_, err := installer.Install(installationManifest)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleaseCompiler.CompileInputs[0].Deployment).To(Equal(installationManifest))
		})

		It("installs the deployment jobs", func() {
			_, err := installer.Install(installationManifest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeJobInstaller.JobInstallInputs).To(Equal(
				[]fakebminstalljob.JobInstallInput{
					{Job: releaseJob},
				},
			))
		})

		It("returns the installation", func() {
			installation, err := installer.Install(installationManifest)
			Expect(err).NotTo(HaveOccurred())

			expectedInstallation := NewInstallation(
				target,
				installedJob,
				installationManifest,
				mockRegistryServerManager,
			)

			Expect(installation).To(Equal(expectedInstallation))
		})

		Context("when the release does not contain a 'cpi' job", func() {
			BeforeEach(func() {
				releaseJob.Name = "not-cpi"
			})

			It("returns an error", func() {
				_, err := installer.Install(installationManifest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid CPI release: job 'cpi' not found in release 'fake-release-name'"))
			})
		})

		Context("when compilation fails", func() {
			JustBeforeEach(func() {
				fakeReleaseCompiler.SetCompileBehavior(release, installationManifest, errors.New("fake-compile-error"))
			})

			It("returns an error", func() {
				_, err := installer.Install(installationManifest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
				Expect(fakeUI.Errors).To(ContainElement("Could not compile CPI release"))
			})
		})

		Context("when the release specified in the manifest cannot be found", func() {
			JustBeforeEach(func() {
				expectFindByName.Return(nil, false)
			})

			It("returns an error", func() {
				_, err := installer.Install(installationManifest)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("CPI release 'fake-release-name' not found"))
				Expect(fakeUI.Errors).To(ContainElement("Could not find CPI release 'fake-release-name'"))
			})
		})
	})
})
