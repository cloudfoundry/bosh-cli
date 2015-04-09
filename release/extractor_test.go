package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/release"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmreljob "github.com/cloudfoundry/bosh-init/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-init/release/pkg"

	fakebmrel "github.com/cloudfoundry/bosh-init/release/fakes"
	testfakes "github.com/cloudfoundry/bosh-init/testutils/fakes"
)

var _ = Describe("Extractor", func() {

	var (
		fakeFS               *fakesys.FakeFileSystem
		fakeExtractor        *testfakes.FakeMultiResponseExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator

		deploymentManifestPath string
		releaseExtractor       Extractor
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		logger := boshlog.NewLogger(boshlog.LevelNone)

		deploymentManifestPath = "/fake/manifest.yml"
		releaseExtractor = NewExtractor(fakeFS, fakeExtractor, fakeReleaseValidator, logger)
	})

	Describe("Extract", func() {
		var (
			releaseTarballPath string
		)
		BeforeEach(func() {
			releaseTarballPath = "/fake/release.tgz"
			fakeFS.WriteFileString(releaseTarballPath, "fake-tgz-contents")
		})

		Context("when a extracted release directory can be created", func() {
			var (
				release    Release
				releaseJob bmreljob.Job
			)

			BeforeEach(func() {
				fakeFS.TempDirDirs = []string{"/extracted-release-path"}

				releasePackage := &bmrelpkg.Package{
					Name:          "fake-release-package-name",
					Fingerprint:   "fake-release-package-fingerprint",
					SHA1:          "fake-release-package-sha1",
					Dependencies:  []*bmrelpkg.Package{},
					ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name",
				}

				releaseJob = bmreljob.Job{
					Name:          "cpi",
					Fingerprint:   "fake-release-job-fingerprint",
					SHA1:          "fake-release-job-sha1",
					ExtractedPath: "/extracted-release-path/extracted_jobs/cpi",
					Templates: map[string]string{
						"cpi.erb":     "bin/cpi",
						"cpi.yml.erb": "config/cpi.yml",
					},
					PackageNames: []string{releasePackage.Name},
					Packages:     []*bmrelpkg.Package{releasePackage},
					Properties:   map[string]bmreljob.PropertyDefinition{},
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
- name: cpi
  version: fake-release-job-version
  fingerprint: fake-release-job-fingerprint
  sha1: fake-release-job-sha1
`
				fakeFS.WriteFileString("/extracted-release-path/release.MF", releaseContents)
				jobManifestContents := `---
name: cpi
templates:
  cpi.erb: bin/cpi
  cpi.yml.erb: config/cpi.yml

packages:
- fake-release-package-name

properties: {}
`
				fakeFS.WriteFileString("/extracted-release-path/extracted_jobs/cpi/job.MF", jobManifestContents)
			})

			JustBeforeEach(func() {
				releaseJobs := []bmreljob.Job{releaseJob}
				releasePackages := append([]*bmrelpkg.Package(nil), releaseJob.Packages...)
				release = NewRelease(
					"fake-release-name",
					"fake-release-version",
					releaseJobs,
					releasePackages,
					"/extracted-release-path",
					fakeFS,
				)
			})

			Context("and the tarball is a valid BOSH release", func() {
				It("extracts the release to the ExtractedPath", func() {
					release, err := releaseExtractor.Extract(releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())

					expectedPackage := &bmrelpkg.Package{
						Name:          "fake-release-package-name",
						Fingerprint:   "fake-release-package-fingerprint",
						SHA1:          "fake-release-package-sha1",
						ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name",
						ArchivePath:   "/extracted-release-path/packages/fake-release-package-name.tgz",
						Dependencies:  []*bmrelpkg.Package{},
					}
					expectedRelease := NewRelease(
						"fake-release-name",
						"fake-release-version",
						[]bmreljob.Job{
							{
								Name:          "cpi",
								Fingerprint:   "fake-release-job-fingerprint",
								SHA1:          "fake-release-job-sha1",
								ExtractedPath: "/extracted-release-path/extracted_jobs/cpi",
								Templates: map[string]string{
									"cpi.erb":     "bin/cpi",
									"cpi.yml.erb": "config/cpi.yml",
								},
								PackageNames: []string{
									"fake-release-package-name",
								},
								Packages:   []*bmrelpkg.Package{expectedPackage},
								Properties: map[string]bmreljob.PropertyDefinition{},
							},
						},
						[]*bmrelpkg.Package{expectedPackage},
						"/extracted-release-path",
						fakeFS,
					)

					Expect(release).To(Equal(expectedRelease))

					Expect(fakeFS.FileExists("/extracted-release-path")).To(BeTrue())
					Expect(fakeFS.FileExists("/extracted-release-path/extracted_packages/fake-release-package-name")).To(BeTrue())
					Expect(fakeFS.FileExists("/extracted-release-path/extracted_jobs/cpi")).To(BeTrue())
				})
			})

			Context("and the tarball is not a valid BOSH release", func() {
				BeforeEach(func() {
					fakeFS.WriteFileString("/extracted-release-path/release.MF", `{}`)
					fakeReleaseValidator.ValidateError = bosherr.Error("fake-error")
				})

				It("returns an error", func() {
					_, err := releaseExtractor.Extract(releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			Context("and the tarball cannot be read", func() {
				It("returns an error", func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/extracted-release-path", bosherr.Error("fake-error"))
					_, err := releaseExtractor.Extract(releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading release from '/fake/release.tgz'"))
				})
			})
		})

		Context("when a extracted release path cannot be created", func() {
			BeforeEach(func() {
				fakeFS.TempDirError = bosherr.Error("fake-tmp-dir-error")
			})

			It("returns an error", func() {
				_, err := releaseExtractor.Extract(releaseTarballPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-tmp-dir-error"))
				Expect(err.Error()).To(ContainSubstring("Creating temp directory to extract release '/fake/release.tgz'"))
			})
		})
	})
})
