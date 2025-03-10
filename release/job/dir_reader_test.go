package job_test

import (
	"errors"
	"path/filepath"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	fakeres "github.com/cloudfoundry/bosh-cli/v7/release/resource/resourcefakes"
)

var _ = Describe("DirReaderImpl", func() {
	var (
		collectedFiles          []File
		collectedPrepFiles      []File
		collectedChunks         []string
		collectedFollowSymlinks bool
		archive                 *fakeres.FakeArchive
		fs                      *fakesys.FakeFileSystem
		reader                  DirReaderImpl
	)

	BeforeEach(func() {
		archive = &fakeres.FakeArchive{}
		archiveFactory := func(args ArchiveFactoryArgs) Archive {
			collectedFiles = args.Files
			collectedPrepFiles = args.PrepFiles
			collectedChunks = args.Chunks
			collectedFollowSymlinks = args.FollowSymlinks
			return archive
		}
		fs = fakesys.NewFakeFileSystem()
		reader = NewDirReaderImpl(archiveFactory, fs)
	})

	Describe("Read", func() {
		It("returns a job with the details collected from job directory", func() {
			err := fs.WriteFileString(filepath.Join("/", "my-job", "spec"), `---
name: my-job
templates: {src: dst}
packages: [pkg]
properties:
  prop:
    description: prop-desc
    default: prop-default
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join("/", "my-job", "monit"), "monit-content")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join("/", "my-job", "properties_schema.json"), "properties-schema-content")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join("/", "my-job", "templates", "src"), "tpl-content")
			Expect(err).ToNot(HaveOccurred())

			archive.FingerprintReturns("fp", nil)

			expectedJob := NewJob(NewResource("my-job", "fp", archive))
			expectedJob.PackageNames = []string{"pkg"} // only expect pkg names

			job, err := reader.Read(filepath.Join("/", "my-job"))
			Expect(err).NotTo(HaveOccurred())
			Expect(job).To(Equal(expectedJob))

			Expect(collectedFiles).To(Equal([]File{
				{Path: filepath.Join("/", "my-job", "spec"), DirPath: filepath.Join("/", "my-job"), RelativePath: "job.MF"},
				{Path: filepath.Join("/", "my-job", "monit"), DirPath: filepath.Join("/", "my-job"), RelativePath: "monit"},
				{Path: filepath.Join("/", "my-job", "properties_schema.json"), DirPath: filepath.Join("/", "my-job"), RelativePath: "properties_schema.json"},
				{Path: filepath.Join("/", "my-job", "templates", "src"), DirPath: filepath.Join("/", "my-job"), RelativePath: filepath.Join("templates", "src")},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
			Expect(collectedFollowSymlinks).To(BeTrue())
		})

		It("returns a job with the details without monit file", func() {
			err := fs.WriteFileString(filepath.Join("/", "my-job", "spec"), "---\nname: my-job")
			Expect(err).ToNot(HaveOccurred())

			archive.FingerprintReturns("fp", nil)

			job, err := reader.Read(filepath.Join("/", "my-job"))
			Expect(err).NotTo(HaveOccurred())
			Expect(job).To(Equal(NewJob(NewResource("my-job", "fp", archive))))

			Expect(collectedFiles).To(Equal([]File{
				{Path: filepath.Join("/", "my-job", "spec"), DirPath: filepath.Join("/", "my-job"), RelativePath: "job.MF"},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("returns error if spec file is not valid", func() {
			err := fs.WriteFileString(filepath.Join("/", "my-job", "spec"), `-`)
			Expect(err).ToNot(HaveOccurred())

			_, err = reader.Read(filepath.Join("/", "my-job"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Collecting job files"))
		})

		It("returns error if fingerprinting fails", func() {
			err := fs.WriteFileString(filepath.Join("/", "my-job", "spec"), "---\nname: my-job")
			Expect(err).ToNot(HaveOccurred())

			archive.FingerprintReturns("", errors.New("fake-err"))

			_, err = reader.Read(filepath.Join("/", "my-job"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if directory name does not match the job name in spec file", func() {
			err := fs.WriteFileString(filepath.Join("/", "my-job-name", "spec"), `---
name: other-job-name
templates: {src: dst}
packages: [pkg]
properties:
  prop:
    description: prop-desc
    default: prop-default
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join("/", "my-job-name", "monit"), "monit-content")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join("/", "my-job-name", "templates", "src"), "tpl-content")
			Expect(err).ToNot(HaveOccurred())

			_, err = reader.Read(filepath.Join("/", "my-job-name"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Job directory 'my-job-name' does not match job name 'other-job-name' in spec"))
		})
	})
})
