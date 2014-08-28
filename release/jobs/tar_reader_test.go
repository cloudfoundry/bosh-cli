package jobs_test

import (
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
)

var _ = Describe("TarReader", func() {
	var (
		fakeExtractor *testfakes.FakeMultiResponseExtractor
		fakeFs        *fakesys.FakeFileSystem
		reader        Reader
	)
	BeforeEach(func() {
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeFs = fakesys.NewFakeFileSystem()
		reader = NewTarReader("/some/job/archive", "/extracted/job", fakeExtractor, fakeFs)
	})

	Context("when the job archive is a valid tar", func() {
		BeforeEach(func() {
			fakeExtractor.SetDecompressBehavior("/some/job/archive", "/extracted/job", nil)
		})

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
		It("returns error", func() {
			_, err := reader.Read()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Extracting job archive `/some/job/archive'"))
		})
	})
})
