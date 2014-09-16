package release_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("tarReader", func() {
	var (
		reader        Reader
		fakeFs        *fakesys.FakeFileSystem
		fakeExtractor *testfakes.FakeMultiResponseExtractor
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		reader = NewReader("/some/release.tgz", "/extracted/release", fakeFs, fakeExtractor)
	})

	Describe("Read", func() {
		Context("when the given release archive is a valid tar", func() {
			BeforeEach(func() {
				fakeExtractor.SetDecompressBehavior("/some/release.tgz", "/extracted/release", nil)
			})

			Context("when the release manifest is valid", func() {
				BeforeEach(func() {
					fakeFs.WriteFileString(
						"/extracted/release/release.MF",
						`---
name: fake-release
version: fake-version

commit_hash: abc123
uncommitted_changes: true

jobs:
- name: fake-job
  version: fake-job-version
  fingerprint: fake-job-fingerprint
  sha1: fake-job-sha

packages:
- name: fake-package
  version: fake-package-version
  fingerprint: fake-package-fingerprint
  sha1: fake-package-sha
  dependencies:
  - fake-package-1
`,
					)
				})

				Context("when the jobs and packages in the release are valid", func() {
					BeforeEach(func() {
						fakeExtractor.SetDecompressBehavior("/extracted/release/jobs/fake-job.tgz", "/extracted/release/extracted_jobs/fake-job", nil)
						fakeFs.WriteFileString(
							"/extracted/release/extracted_jobs/fake-job/job.MF",
							`---
name: fake-job
templates:
  some_template: some_file
packages:
- fake-package
`,
						)
					})

					Context("when the packages in the release are valid", func() {
						BeforeEach(func() {
							fakeExtractor.SetDecompressBehavior(
								"/extracted/release/packages/fake-package.tgz",
								"/extracted/release/extracted_packages/fake-package",
								nil,
							)
						})

						It("returns a release from the given tar file", func() {
							release, err := reader.Read()
							Expect(err).NotTo(HaveOccurred())
							Expect(release.Name).To(Equal("fake-release"))
							Expect(release.Version).To(Equal("fake-version"))
							Expect(release.CommitHash).To(Equal("abc123"))
							Expect(release.UncommittedChanges).To(BeTrue())
							Expect(release.ExtractedPath).To(Equal("/extracted/release"))
							expectedPackage := &Package{
								Name:          "fake-package",
								Fingerprint:   "fake-package-fingerprint",
								SHA1:          "fake-package-sha",
								Dependencies:  []*Package{&Package{Name: "fake-package-1"}},
								ExtractedPath: "/extracted/release/extracted_packages/fake-package",
							}

							Expect(len(release.Jobs)).To(Equal(1))
							Expect(release.Jobs).To(
								ContainElement(
									Job{
										Name:          "fake-job",
										Fingerprint:   "fake-job-fingerprint",
										SHA1:          "fake-job-sha",
										ExtractedPath: "/extracted/release/extracted_jobs/fake-job",
										Templates:     map[string]string{"some_template": "some_file"},
										PackageNames:  []string{"fake-package"},
										Packages:      []*Package{expectedPackage},
									},
								),
							)

							Expect(len(release.Packages)).To(Equal(1))
							Expect(release.Packages).To(ContainElement(expectedPackage))
						})
					})

					Context("when the package cannot be extracted", func() {
						BeforeEach(func() {
							fakeExtractor.SetDecompressBehavior("/some/release.tgz", "/extracted/release", errors.New("Extracting package `fake-package'"))
						})

						It("returns errors for each invalid package", func() {
							_, err := reader.Read()
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Extracting package `fake-package'"))
						})
					})
				})

				Context("when the jobs in the release are not valid", func() {
					BeforeEach(func() {
						fakeFs.WriteFileString(
							"/extracted/release/release.MF",
							`---
name: fake-release
version: fake-version

jobs:
- name: fake-job
  version: fake-job-version
  fingerprint: fake-job-fingerprint
  sha1: fake-job-sha
- name: fake-job-2
  version: fake-job-2-version
  fingerprint: fake-job-2-fingerprint
  sha1: fake-job-2-sha

packages:
- name: fake-package
  version: fake-package-version
  fingerprint: fake-package-fingerprint
  sha1: fake-package-sha
  dependencies:
  - fake-package-1
`,
						)
					})

					It("returns errors for each invalid job", func() {
						_, err := reader.Read()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Reading job `fake-job' from archive"))
						Expect(err.Error()).To(ContainSubstring("Reading job `fake-job-2' from archive"))
					})
				})

				Context("when an extracted job path cannot be created", func() {
					BeforeEach(func() {
						fakeFs.MkdirAllError = errors.New("")
					})

					It("returns err", func() {
						_, err := reader.Read()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Creating extracted job path"))
					})
				})
			})

			Context("when the CPI release manifest is invalid", func() {
				BeforeEach(func() {
					fakeFs.WriteFileString("/extracted/release/release.MF", "{")
				})

				It("returns an error when the YAML in unparseable", func() {
					_, err := reader.Read()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Parsing release manifest"))
				})

				It("returns an error when the release manifest is missing", func() {
					fakeFs.RemoveAll("/extracted/release/release.MF")
					_, err := reader.Read()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading release manifest"))
				})
			})

			Context("when the job refers to a package that does not exist", func() {
				It("returns error", func() {
					releaseMFContents :=
						`---
name: fake-release
version: fake-version

commit_hash: abc123
uncommitted_changes: true

jobs:
- name: fake-job
version: fake-job-version
fingerprint: fake-job-fingerprint
sha1: fake-job-sha
`
					fakeFs.WriteFileString("/extracted/release/release.MF", releaseMFContents)
					fakeExtractor.SetDecompressBehavior("/extracted/release/jobs/fake-job.tgz", "/extracted/release/extracted_jobs/fake-job", nil)
					jobMFContents :=
						`---
name: fake-job
templates:
  some_template: some_file
packages:
- not_there
`
					fakeFs.WriteFileString("/extracted/release/extracted_jobs/fake-job/job.MF", jobMFContents)
					_, err := reader.Read()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Package not found"))
				})
			})
		})

		Context("when the CPI release is not a valid tar", func() {
			It("returns err", func() {
				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Extracting release"))
			})
		})
	})
})
