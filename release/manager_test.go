package release_test

import (
	"fmt"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
)

var _ = Describe("Manager", func() {

	var (
		fakeFS               *fakesys.FakeFileSystem
		fakeExtractor        *testfakes.FakeMultiResponseExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator

		deploymentManifestPath string
		releaseManager         Manager
	)

	var allowReleaseToBeExtracted = func(name, version, tarballPath string) {
		fakeFS.WriteFileString(tarballPath, "ignored-tgz-contents")

		extractedPath := fmt.Sprintf("/extracted-release-path-%s-%s", name, version)
		fakeFS.TempDirDirs = append(fakeFS.TempDirDirs, extractedPath)

		releaseContentsA := fmt.Sprintf(`---
name: %s
version: %s

packages: []
jobs: []
`, name, version)
		fakeFS.WriteFileString(path.Join(extractedPath, "release.MF"), releaseContentsA)
	}

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		logger := boshlog.NewLogger(boshlog.LevelNone)

		deploymentManifestPath = "/fake/manifest.yml"
		releaseManager = NewManager(fakeFS, fakeExtractor, fakeReleaseValidator, logger)
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
				releaseJob Job
			)

			BeforeEach(func() {
				fakeFS.TempDirDirs = []string{"/extracted-release-path"}

				releasePackage := &Package{
					Name:          "fake-release-package-name",
					Fingerprint:   "fake-release-package-fingerprint",
					SHA1:          "fake-release-package-sha1",
					Dependencies:  []*Package{},
					ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name",
				}

				releaseJob = Job{
					Name:          "cpi",
					Fingerprint:   "fake-release-job-fingerprint",
					SHA1:          "fake-release-job-sha1",
					ExtractedPath: "/extracted-release-path/extracted_jobs/cpi",
					Templates: map[string]string{
						"cpi.erb":     "bin/cpi",
						"cpi.yml.erb": "config/cpi.yml",
					},
					PackageNames: []string{releasePackage.Name},
					Packages:     []*Package{releasePackage},
					Properties:   map[string]PropertyDefinition{},
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
				releaseJobs := []Job{releaseJob}
				releasePackages := append([]*Package(nil), releaseJob.Packages...)
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
					release, err := releaseManager.Extract(releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())

					expectedPackage := &Package{
						Name:          "fake-release-package-name",
						Fingerprint:   "fake-release-package-fingerprint",
						SHA1:          "fake-release-package-sha1",
						ExtractedPath: "/extracted-release-path/extracted_packages/fake-release-package-name",
						Dependencies:  []*Package{},
					}
					expectedRelease := NewRelease(
						"fake-release-name",
						"fake-release-version",
						[]Job{
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
								Packages:   []*Package{expectedPackage},
								Properties: map[string]PropertyDefinition{},
							},
						},
						[]*Package{expectedPackage},
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
					_, err := releaseManager.Extract(releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			Context("and the tarball cannot be read", func() {
				It("returns an error", func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/extracted-release-path", bosherr.Error("fake-error"))
					_, err := releaseManager.Extract(releaseTarballPath)
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
				_, err := releaseManager.Extract(releaseTarballPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-tmp-dir-error"))
				Expect(err.Error()).To(ContainSubstring("Creating temp directory to extract release '/fake/release.tgz'"))
			})
		})
	})

	Describe("List", func() {
		var (
			releaseTarballPathA = "/fake/release-a.tgz"
			releaseTarballPathB = "/fake/release-b.tgz"
		)

		BeforeEach(func() {
			fakeFS.TempDirDirs = []string{}
			allowReleaseToBeExtracted("release-a", "1.0", releaseTarballPathA)
			allowReleaseToBeExtracted("release-b", "1.1", releaseTarballPathB)
		})

		It("returns all releases that have been extracted", func() {
			releaseA, err := releaseManager.Extract(releaseTarballPathA)
			Expect(err).NotTo(HaveOccurred())

			releaseB, err := releaseManager.Extract(releaseTarballPathB)
			Expect(err).NotTo(HaveOccurred())

			Expect(releaseManager.List()).To(Equal([]Release{releaseA, releaseB}))
		})
	})

	Describe("Find", func() {
		It("returns false when no releases have been extracted", func() {
			_, found := releaseManager.Find("release-a")
			Expect(found).To(BeFalse())
		})

		Context("when releases have been extracted", func() {
			var (
				releaseTarballPathA = "/fake/release-a.tgz"
				releaseTarballPathB = "/fake/release-b.tgz"
			)

			BeforeEach(func() {
				fakeFS.TempDirDirs = []string{}
				allowReleaseToBeExtracted("release-a", "1.0", releaseTarballPathA)
				allowReleaseToBeExtracted("release-b", "1.1", releaseTarballPathB)
			})

			It("returns true and the release with the requested name", func() {
				releaseA, err := releaseManager.Extract(releaseTarballPathA)
				Expect(err).NotTo(HaveOccurred())

				releaseB, err := releaseManager.Extract(releaseTarballPathB)
				Expect(err).NotTo(HaveOccurred())

				releaseAFound, found := releaseManager.Find("release-a")
				Expect(found).To(BeTrue())
				Expect(releaseAFound).To(Equal(releaseA))

				releaseBFound, found := releaseManager.Find("release-b")
				Expect(found).To(BeTrue())
				Expect(releaseBFound).To(Equal(releaseB))
			})

			It("returns false when the requested release has not been extracted", func() {
				_, err := releaseManager.Extract(releaseTarballPathA)
				Expect(err).NotTo(HaveOccurred())

				_, found := releaseManager.Find("release-c")
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("DeleteAll", func() {
		var (
			releaseTarballPathA = "/fake/release-a.tgz"
			releaseTarballPathB = "/fake/release-b.tgz"
		)

		BeforeEach(func() {
			fakeFS.TempDirDirs = []string{}
			allowReleaseToBeExtracted("release-a", "1.0", releaseTarballPathA)
			allowReleaseToBeExtracted("release-b", "1.1", releaseTarballPathB)
		})

		It("deletes all extracted releases", func() {
			releaseA, err := releaseManager.Extract(releaseTarballPathA)
			Expect(err).NotTo(HaveOccurred())

			releaseB, err := releaseManager.Extract(releaseTarballPathB)
			Expect(err).NotTo(HaveOccurred())

			err = releaseManager.DeleteAll()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseManager.List()).To(BeEmpty())
			Expect(releaseA.Exists()).To(BeFalse())
			Expect(releaseB.Exists()).To(BeFalse())
		})
	})
})
