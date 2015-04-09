package templatescompiler_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-init/templatescompiler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmindex "github.com/cloudfoundry/bosh-init/index"
	bmreljob "github.com/cloudfoundry/bosh-init/release/job"
)

var _ = Describe("TemplatesRepo", func() {
	var (
		index         bmindex.Index
		templatesRepo TemplatesRepo
		fakeFS        *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		index = bmindex.NewFileIndex("/index_file", fakeFS)
		templatesRepo = NewTemplatesRepo(index)
	})

	Context("Save and Find", func() {
		var (
			record TemplateRecord
			job    bmreljob.Job
		)

		BeforeEach(func() {
			record = TemplateRecord{}
			job = bmreljob.Job{
				Name:        "fake-job-name",
				Fingerprint: "fake-job-fingerprint",
			}
		})

		It("saves the rendered template to the index", func() {
			err := templatesRepo.Save(job, record)
			Expect(err).ToNot(HaveOccurred())

			result, found, err := templatesRepo.Find(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(result).To(Equal(record))
		})

		It("returns false when finding before saving", func() {
			_, found, err := templatesRepo.Find(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		It("returns false if job fingerprint have changed after saving", func() {
			err := templatesRepo.Save(job, record)
			Expect(err).ToNot(HaveOccurred())

			_, found, err := templatesRepo.Find(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			job.Fingerprint = "new-fake-job-fingerprint"

			_, found, err = templatesRepo.Find(job)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		Context("when saving to index fails", func() {
			It("returns error", func() {
				fakeFS.WriteToFileError = errors.New("fake-write-error")
				record := TemplateRecord{
					BlobID:   "fake-blob-id",
					BlobSHA1: "fake-sha1",
				}

				err := templatesRepo.Save(job, record)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-write-error"))
			})
		})

		Context("when reading from index fails", func() {
			It("returns error", func() {
				err := templatesRepo.Save(job, record)
				fakeFS.ReadFileError = errors.New("fake-read-error")

				_, _, err = templatesRepo.Find(job)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-read-error"))
			})
		})
	})
})
