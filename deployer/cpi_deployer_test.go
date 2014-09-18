package deployer_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/install"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakebmjobi "github.com/cloudfoundry/bosh-micro-cli/install/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("CpiDeployer", func() {

	var (
		fakeFs               *fakesys.FakeFileSystem
		fakeExtractor        *testfakes.FakeMultiResponseExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator
		fakeReleaseCompiler  *fakebmcomp.FakeReleaseCompiler
		fakeJobInstaller     *fakebmjobi.FakeJobInstaller
		fakeCloudFactory     *fakebmcloud.FakeFactory
		fakeUI               *fakebmui.FakeUI

		deploymentManifestPath string
		cpiDeployer            CpiDeployer
	)
	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		fakeReleaseCompiler = fakebmcomp.NewFakeReleaseCompiler()
		fakeJobInstaller = fakebmjobi.NewFakeJobInstaller()
		fakeCloudFactory = fakebmcloud.NewFakeFactory()
		fakeUI = &fakebmui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)

		deploymentManifestPath = "/fake/manifest.yml"
		cpiDeployer = NewCpiDeployer(fakeUI, fakeFs, fakeExtractor, fakeReleaseValidator, fakeReleaseCompiler, fakeJobInstaller, fakeCloudFactory, logger)
	})

	Describe("Deploy", func() {
		var (
			releaseTarballPath string
			deployment         bmdepl.Deployment
		)
		BeforeEach(func() {
			fakeFs.WriteFileString(deploymentManifestPath, "")

			releaseTarballPath = "/fake/release.tgz"

			fakeFs.WriteFileString(releaseTarballPath, "")
			deployment = bmdepl.Deployment{}
		})

		Context("when a extracted release directory can be created", func() {
			var (
				release    bmrel.Release
				releaseJob bmrel.Job
			)

			BeforeEach(func() {
				fakeFs.TempDirDir = "/release"

				releasePackage := &bmrel.Package{
					Name:          "fake-release-package-name",
					Fingerprint:   "fake-release-package-fingerprint",
					SHA1:          "fake-release-package-sha1",
					Dependencies:  []*bmrel.Package{},
					ExtractedPath: "/release/extracted_packages/fake-release-package-name",
				}

				releaseJob = bmrel.Job{
					Name:          "fake-release-job-name",
					Fingerprint:   "fake-release-job-fingerprint",
					SHA1:          "fake-release-job-sha1",
					ExtractedPath: "/release/extracted_jobs/fake-release-job-name",
					Templates: map[string]string{
						"cpi.erb":               "bin/cpi",
						"micro_discover_ip.erb": "bin/micro_discover_ip",
					},
					PackageNames: []string{releasePackage.Name},
					Packages:     []*bmrel.Package{releasePackage},
					Properties:   map[string]bmrel.PropertyDefinition{},
				}

				release = bmrel.Release{
					Name:          "fake-release-name",
					Version:       "fake-release-version",
					Jobs:          []bmrel.Job{releaseJob},
					Packages:      []*bmrel.Package{releasePackage},
					ExtractedPath: "/release",
					TarballPath:   releaseTarballPath,
				}

				releaseContents := `---
name: fake-release-name
version: fake-release-version

packages:
- name: fake-release-package-name
  version: fake-release-package-version
  fingerprint: fake-release-package-fingerprint
  sha1: fake-release-package-sha1
  dependencies: []
jobs:
- name: fake-release-job-name
  version: fake-release-job-version
  fingerprint: fake-release-job-fingerprint
  sha1: fake-release-job-sha1
`
				fakeFs.WriteFileString("/release/release.MF", releaseContents)
				fakeExtractor.SetDecompressBehavior("/release/packages/fake-release-package-name.tgz", "/release/extracted_packages/fake-release-package-name", nil)
				fakeExtractor.SetDecompressBehavior("/release/jobs/fake-release-job-name.tgz", "/release/extracted_jobs/fake-release-job-name", nil)
				jobManifestContents := `---
name: fake-release-job-name
templates:
 cpi.erb: bin/cpi
 micro_discover_ip.erb: bin/micro_discover_ip

packages:
 - fake-release-package-name

properties: {}
`
				fakeFs.WriteFileString("/release/extracted_jobs/fake-release-job-name/job.MF", jobManifestContents)
			})

			Context("and the tarball is a valid BOSH release", func() {
				var (
					installedJob  bminstall.InstalledJob
					installedJobs []bminstall.InstalledJob
					cloud         *fakebmcloud.FakeCloud
				)

				BeforeEach(func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/release", nil)

					//TODO: parse deployment from yml?
					deployment = bmdepl.Deployment{
						Name:       "fake-deployment-name",
						Properties: map[string]interface{}{},
						Jobs: []bmdepl.Job{
							bmdepl.Job{
								Name:      "fake-deployment-job-name",
								Instances: 1,
								Templates: []bmdepl.ReleaseJobRef{
									bmdepl.ReleaseJobRef{
										Name:    "fake-release-job-name",
										Release: "fake-release-name",
									},
								},
							},
						},
					}

					fakeReleaseCompiler.SetCompileBehavior(release, deployment, nil)

					installedJob = bminstall.InstalledJob{
						Name: "fake-release-job-name",
						Path: "/release/fake-release-job-name",
					}
					fakeJobInstaller.SetInstallBehavior(releaseJob, installedJob, nil)

					installedJobs = []bminstall.InstalledJob{installedJob}
					cloud = fakebmcloud.NewFakeCloud()
					fakeCloudFactory.SetNewCloudBehavior(installedJobs, cloud, nil)
				})

				It("does not return an error", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("compiles the release", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeReleaseCompiler.CompileInputs[0].Deployment).To(Equal(deployment))
				})

				It("cleans up the extracted release directory", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeFs.FileExists("/release")).To(BeFalse())
				})

				It("installs the deployment jobs", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeJobInstaller.JobInstallInputs).To(Equal(
						[]fakebmjobi.JobInstallInput{
							fakebmjobi.JobInstallInput{
								Job: releaseJob,
							},
						},
					))
				})
			})

			Context("and the tarball is not a valid BOSH release", func() {
				BeforeEach(func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/release", nil)
					fakeFs.WriteFileString("/release/release.MF", `{}`)
					fakeReleaseValidator.ValidateError = errors.New("fake-error")
				})

				It("returns an error", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})

				It("cleans up the extracted release directory", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(fakeFs.FileExists("/release")).To(BeFalse())
				})
			})

			Context("and the tarball cannot be read", func() {
				It("returns an error", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading CPI release from `/fake/release.tgz'"))
					Expect(fakeUI.Errors).To(ContainElement("CPI release at `/fake/release.tgz' is not a BOSH release"))
				})
			})

			Context("when compilation fails", func() {
				It("returns an error", func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/release", nil)
					fakeReleaseCompiler.SetCompileBehavior(release, deployment, errors.New("fake-compile-error"))

					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
					Expect(fakeUI.Errors).To(ContainElement("Could not compile CPI release"))
				})
			})
		})

		Context("when a extracted release path cannot be created", func() {
			BeforeEach(func() {
				fakeFs.TempDirError = errors.New("fake-tmp-dir-error")
			})

			It("returns an error", func() {
				_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-tmp-dir-error"))
				Expect(err.Error()).To(ContainSubstring("Creating temp directory"))
				Expect(fakeUI.Errors).To(ContainElement("Could not create a temporary directory"))
			})
		})
	})
})
