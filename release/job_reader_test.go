package release_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("JobReader", func() {
	var (
		fakeExtractor *testfakes.FakeMultiResponseExtractor
		fakeFs        *fakesys.FakeFileSystem
		reader        JobReader
	)
	BeforeEach(func() {
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeFs = fakesys.NewFakeFileSystem()
		reader = NewJobReader("/some/job/archive", "/extracted/job", fakeExtractor, fakeFs)
	})

	Context("when the job archive is a valid tar", func() {
		Context("when the job manifest is valid", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(
					"/extracted/job/job.MF",
					`---
name: fake-job
templates:
  some_template: some_file
packages:
- fake-package
properties:
  fake-property:
    description: "Fake description"
    default: "fake-default"
`,
				)
			})

			It("returns a job with the details from the manifest", func() {
				job, err := reader.Read()
				Expect(err).NotTo(HaveOccurred())
				Expect(job).To(Equal(
					Job{
						Name:          "fake-job",
						Templates:     map[string]string{"some_template": "some_file"},
						PackageNames:  []string{"fake-package"},
						ExtractedPath: "/extracted/job",
						Properties: map[string]PropertyDefinition{
							"fake-property": PropertyDefinition{
								Description: "Fake description",
								RawDefault:  "fake-default",
							},
						},
					},
				))
			})
		})

		Context("when the job manifest is invalid", func() {
			It("returns an error when the job manifest is missing", func() {
				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Reading job manifest"))
			})

			It("returns an error when the job manifest is invalid", func() {
				fakeFs.WriteFileString("/extracted/job/job.MF", "{")
				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing job manifest"))
			})
		})
	})

	Context("when the job archive is not a valid tar", func() {
		BeforeEach(func() {
			fakeExtractor.SetDecompressBehavior("/some/job/archive", "/extracted/job", errors.New("fake-error"))
		})

		It("returns error", func() {
			_, err := reader.Read()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Extracting job archive `/some/job/archive'"))
		})
	})
})
