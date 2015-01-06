package cpi_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/cpi"

	"code.google.com/p/gomock/gomock"
	mock_release "github.com/cloudfoundry/bosh-micro-cli/release/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/cpi/compile/fakes"
	fakebmjobi "github.com/cloudfoundry/bosh-micro-cli/cpi/install/fakes"
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
		fakeJobInstaller    *fakebmjobi.FakeJobInstaller
		fakeCloudFactory    *fakebmcloud.FakeFactory
		fakeUI              *fakebmui.FakeUI

		mockReleaseManager *mock_release.MockManager

		deploymentManifestPath string
		cpiInstaller           Installer
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseCompiler = fakebmcomp.NewFakeReleaseCompiler()
		fakeJobInstaller = fakebmjobi.NewFakeJobInstaller()
		fakeCloudFactory = fakebmcloud.NewFakeFactory()
		fakeUI = &fakebmui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)

		mockReleaseManager = mock_release.NewMockManager(mockCtrl)

		deploymentManifestPath = "/fake/manifest.yml"
		cpiInstaller = NewInstaller(fakeUI, fakeFS, fakeExtractor, mockReleaseManager, fakeReleaseCompiler, fakeJobInstaller, fakeCloudFactory, logger)
	})

	Describe("Install", func() {
		var (
			deployment bminstallmanifest.Manifest
			release    bmrel.Release
			releaseJob bmrel.Job

			directorID = "fake-director-id"

			installedJob   bmcpiinstall.InstalledJob
			installedCloud *fakebmcloud.FakeCloud

			expectFindRelease *gomock.Call
		)

		BeforeEach(func() {
			fakeFS.WriteFileString(deploymentManifestPath, "")

			deployment = bminstallmanifest.Manifest{
				Name: "fake-deployment-name",
				Release: bmdeplmanifest.ReleaseRef{
					Name:    "fake-release-name",
					Version: "fake-release-version",
				},
				RawProperties: map[interface{}]interface{}{},
				Jobs: []bmdeplmanifest.Job{
					{
						Name:      "cpi",
						Instances: 1,
						Templates: []bmdeplmanifest.ReleaseJobRef{
							{
								Name:    "cpi",
								Release: "fake-release-name",
							},
						},
					},
				},
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

			installedJob = bmcpiinstall.InstalledJob{
				Name: "cpi",
				Path: "/extracted-release-path/cpi",
			}

			installedCloud = fakebmcloud.NewFakeCloud()
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

			fakeJobInstaller.SetInstallBehavior(releaseJob, func(_ bmrel.Job) (bmcpiinstall.InstalledJob, error) {
				return installedJob, nil
			})

			fakeCloudFactory.SetNewCloudBehavior(installedJob, directorID, installedCloud, nil)

			fakeReleaseCompiler.SetCompileBehavior(release, deployment, nil)

			fakeFS.MkdirAll("/extracted-release-path", os.FileMode(0750))

			expectFindRelease = mockReleaseManager.EXPECT().Find("fake-release-name", "fake-release-version").Return(release, true)
		})

		It("compiles the release", func() {
			_, err := cpiInstaller.Install(deployment, directorID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleaseCompiler.CompileInputs[0].Deployment).To(Equal(deployment))
		})

		It("installs the deployment jobs", func() {
			_, err := cpiInstaller.Install(deployment, directorID)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeJobInstaller.JobInstallInputs).To(Equal(
				[]fakebmjobi.JobInstallInput{
					{Job: releaseJob},
				},
			))
		})

		It("returns a cloud wrapper around the installed CPI", func() {
			cloud, err := cpiInstaller.Install(deployment, directorID)
			Expect(err).NotTo(HaveOccurred())

			Expect(cloud).To(Equal(installedCloud))
		})

		Context("when the release does not contain a 'cpi' job", func() {
			BeforeEach(func() {
				releaseJob.Name = "not-cpi"
			})

			It("returns an error", func() {
				_, err := cpiInstaller.Install(deployment, directorID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid CPI release: job 'cpi' not found in release 'fake-release-name'"))
			})
		})

		Context("when compilation fails", func() {
			JustBeforeEach(func() {
				fakeReleaseCompiler.SetCompileBehavior(release, deployment, errors.New("fake-compile-error"))
			})

			It("returns an error", func() {
				_, err := cpiInstaller.Install(deployment, directorID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
				Expect(fakeUI.Errors).To(ContainElement("Could not compile CPI release"))
			})
		})

		Context("when the release specified in the manifest cannot be found", func() {
			JustBeforeEach(func() {
				expectFindRelease.Return(nil, false)
			})

			It("returns an error", func() {
				_, err := cpiInstaller.Install(deployment, directorID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("CPI release 'fake-release-name/fake-release-version' not found"))
				Expect(fakeUI.Errors).To(ContainElement("Could not find CPI release 'fake-release-name/fake-release-version'"))
			})
		})
	})
})
