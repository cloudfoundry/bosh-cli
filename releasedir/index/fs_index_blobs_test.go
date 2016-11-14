package index_test

import (
	"errors"
	"os"
	"syscall"

	fakeblob "github.com/cloudfoundry/bosh-utils/blobstore/fakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakecrypto "github.com/cloudfoundry/bosh-cli/crypto/fakes"
	boshidx "github.com/cloudfoundry/bosh-cli/releasedir/index"
	fakeidx "github.com/cloudfoundry/bosh-cli/releasedir/index/indexfakes"
)

var _ = Describe("FSIndexBlobs", func() {
	var (
		reporter  *fakeidx.FakeReporter
		blobstore *fakeblob.FakeBlobstore
		sha1calc  *fakecrypto.FakeSha1Calculator
		fs        *fakesys.FakeFileSystem
		blobs     boshidx.FSIndexBlobs
	)

	BeforeEach(func() {
		reporter = &fakeidx.FakeReporter{}
		blobstore = nil
		sha1calc = fakecrypto.NewFakeSha1Calculator()
		fs = fakesys.NewFakeFileSystem()
	})

	Describe("Get", func() {
		itChecksIfFileIsAlreadyDownloaded := func() {
			Context("when local copy exists", func() {
				BeforeEach(func() {
					sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
						"/dir/sub-dir/sha1": fakecrypto.CalculateInput{Sha1: "sha1"},
						"/full-dir/sha1":    fakecrypto.CalculateInput{Sha1: "sha1"},
					})
				})

				It("returns path to a downloaded blob if it already exists", func() {
					fs.WriteFileString("/dir/sub-dir/sha1", "file")

					path, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).ToNot(HaveOccurred())
					Expect(path).To(Equal("/dir/sub-dir/sha1"))
				})

				It("returns error if local copy not match expected sha1", func() {
					sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
						"/dir/sub-dir/sha1": fakecrypto.CalculateInput{Sha1: "wrong-sha1"},
					})
					fs.WriteFileString("/dir/sub-dir/sha1", "file")

					_, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(
						"Expected local copy ('/dir/sub-dir/sha1') of blob 'blob-id' to have SHA1 'sha1' but was 'wrong-sha1'"))
				})

				It("returns error if cannot check local copy's sha1", func() {
					sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
						"/dir/sub-dir/sha1": fakecrypto.CalculateInput{Err: errors.New("fake-err")},
					})
					fs.WriteFileString("/dir/sub-dir/sha1", "file")

					_, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("expands directory path", func() {
					fs.ExpandPathExpanded = "/full-dir"
					fs.WriteFileString("/full-dir/sha1", "file")

					path, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).ToNot(HaveOccurred())
					Expect(path).To(Equal("/full-dir/sha1"))

					Expect(fs.ExpandPathPath).To(Equal("/dir/sub-dir"))
				})

				It("returns error if expanding directory path fails", func() {
					fs.ExpandPathErr = errors.New("fake-err")

					_, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("returns error if creating directory fails", func() {
					fs.MkdirAllError = errors.New("fake-err")

					_, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})
			})
		}

		Context("when configured without a blobstore", func() {
			BeforeEach(func() {
				blobs = boshidx.NewFSIndexBlobs("/dir/sub-dir", reporter, nil, sha1calc, fs)
			})

			itChecksIfFileIsAlreadyDownloaded()

			It("returns error if downloaded blob does not exist", func() {
				_, err := blobs.Get("name", "blob-id", "sha1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot find blob 'blob-id' with SHA1 'sha1'"))
			})

			It("returns error if blob id is not provided", func() {
				_, err := blobs.Get("name", "", "sha1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot find blob named 'name' with SHA1 'sha1'"))
			})
		})

		Context("when configured with a blobstore", func() {
			BeforeEach(func() {
				blobstore = fakeblob.NewFakeBlobstore()
				blobs = boshidx.NewFSIndexBlobs("/dir/sub-dir", reporter, blobstore, sha1calc, fs)
			})

			itChecksIfFileIsAlreadyDownloaded()

			It("downloads blob and places it into a cache", func() {
				blobstore.GetFileName = "/tmp/downloaded-path"
				fs.WriteFileString("/tmp/downloaded-path", "blob")

				path, err := blobs.Get("name", "blob-id", "sha1")
				Expect(err).ToNot(HaveOccurred())
				Expect(path).To(Equal("/dir/sub-dir/sha1"))

				Expect(fs.ReadFileString("/dir/sub-dir/sha1")).To(Equal("blob"))
				Expect(fs.FileExists("/tmp/downloaded-path")).To(BeFalse())

				Expect(reporter.IndexEntryDownloadStartedCallCount()).To(Equal(1))
				Expect(reporter.IndexEntryDownloadFinishedCallCount()).To(Equal(1))

				kind, desc := reporter.IndexEntryDownloadStartedArgsForCall(0)
				Expect(kind).To(Equal("name"))
				Expect(desc).To(Equal("sha1=sha1"))

				kind, desc, err = reporter.IndexEntryDownloadFinishedArgsForCall(0)
				Expect(kind).To(Equal("name"))
				Expect(desc).To(Equal("sha1=sha1"))
				Expect(err).To(BeNil())
			})

			Context("when downloading blob fails", func() {
				It("returns error", func() {
					blobstore.GetError = errors.New("fake-err")

					_, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
					Expect(err.Error()).To(ContainSubstring("Downloading blob 'blob-id'"))

					Expect(reporter.IndexEntryDownloadStartedCallCount()).To(Equal(1))
					Expect(reporter.IndexEntryDownloadFinishedCallCount()).To(Equal(1))

					kind, desc := reporter.IndexEntryDownloadStartedArgsForCall(0)
					Expect(kind).To(Equal("name"))
					Expect(desc).To(Equal("sha1=sha1"))

					kind, desc, err = reporter.IndexEntryDownloadFinishedArgsForCall(0)
					Expect(kind).To(Equal("name"))
					Expect(desc).To(Equal("sha1=sha1"))
					Expect(err).ToNot(BeNil())
				})
			})

			Context("when moving blob into cache fails for unknown reason", func() {
				It("returns error", func() {
					fs.RenameError = errors.New("fake-err")

					_, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
					Expect(err.Error()).To(ContainSubstring("Moving blob 'blob-id'"))

					Expect(reporter.IndexEntryDownloadStartedCallCount()).To(Equal(1))
					Expect(reporter.IndexEntryDownloadFinishedCallCount()).To(Equal(1))

					kind, desc := reporter.IndexEntryDownloadStartedArgsForCall(0)
					Expect(kind).To(Equal("name"))
					Expect(desc).To(Equal("sha1=sha1"))

					kind, desc, err = reporter.IndexEntryDownloadFinishedArgsForCall(0)
					Expect(kind).To(Equal("name"))
					Expect(desc).To(Equal("sha1=sha1"))
					Expect(err).ToNot(BeNil())
				})
			})

			Context("when moving blob onto separate device", func() {
				BeforeEach(func() {
					fs.RenameError = &os.LinkError{
						Err: syscall.Errno(0x12),
					}
				})

				It("It successfully moves blob", func() {
					blobstore.GetFileName = "/tmp/downloaded-path"
					fs.WriteFileString("/tmp/downloaded-path", "blob")

					path, err := blobs.Get("name", "blob-id", "sha1")
					Expect(err).ToNot(HaveOccurred())
					Expect(path).To(Equal("/dir/sub-dir/sha1"))

					Expect(fs.ReadFileString("/dir/sub-dir/sha1")).To(Equal("blob"))
					Expect(fs.FileExists("/tmp/downloaded-path")).To(BeFalse())

					Expect(reporter.IndexEntryDownloadStartedCallCount()).To(Equal(1))
					Expect(reporter.IndexEntryDownloadFinishedCallCount()).To(Equal(1))

					kind, desc := reporter.IndexEntryDownloadStartedArgsForCall(0)
					Expect(kind).To(Equal("name"))
					Expect(desc).To(Equal("sha1=sha1"))

					kind, desc, err = reporter.IndexEntryDownloadFinishedArgsForCall(0)
					Expect(kind).To(Equal("name"))
					Expect(desc).To(Equal("sha1=sha1"))
					Expect(err).To(BeNil())
				})

				Context("when file copy across devices fails", func() {
					It("returns error", func() {
						fs.CopyFileError = errors.New("copy-err")

						_, err := blobs.Get("name", "blob-id", "sha1")
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("copy-err"))
						Expect(err.Error()).To(ContainSubstring("Moving blob 'blob-id'"))

						Expect(reporter.IndexEntryDownloadStartedCallCount()).To(Equal(1))
						Expect(reporter.IndexEntryDownloadFinishedCallCount()).To(Equal(1))

						kind, desc := reporter.IndexEntryDownloadStartedArgsForCall(0)
						Expect(kind).To(Equal("name"))
						Expect(desc).To(Equal("sha1=sha1"))

						kind, desc, err = reporter.IndexEntryDownloadFinishedArgsForCall(0)
						Expect(kind).To(Equal("name"))
						Expect(desc).To(Equal("sha1=sha1"))
						Expect(err).ToNot(BeNil())
					})
				})
			})

			It("returns error if blob id is not provided", func() {
				_, err := blobs.Get("name", "", "sha1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot find blob named 'name' with SHA1 'sha1'"))
			})
		})
	})

	Describe("Add", func() {
		BeforeEach(func() {
			fs.WriteFileString("/tmp/sha1", "file")
		})

		itCopiesFileIntoDir := func() {
			It("copies file into cache dir", func() {
				blobID, path, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).ToNot(HaveOccurred())
				Expect(blobID).To(Equal(""))
				Expect(path).To(Equal("/dir/sub-dir/sha1"))

				Expect(fs.ReadFileString("/dir/sub-dir/sha1")).To(Equal("file"))
			})

			It("keeps existing file in the cache directory if it's already there", func() {
				fs.WriteFileString("/dir/sub-dir/sha1", "other")

				blobID, path, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).ToNot(HaveOccurred())
				Expect(blobID).To(Equal(""))
				Expect(path).To(Equal("/dir/sub-dir/sha1"))

				Expect(fs.ReadFileString("/dir/sub-dir/sha1")).To(Equal("other"))
			})

			It("expands directory path", func() {
				fs.ExpandPathExpanded = "/full-dir"
				fs.WriteFileString("/full-dir/sha1", "file")

				_, _, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.ExpandPathPath).To(Equal("/dir/sub-dir"))
			})

			It("returns error if expanding directory path fails", func() {
				fs.ExpandPathErr = errors.New("fake-err")

				_, _, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if creating directory fails", func() {
				fs.MkdirAllError = errors.New("fake-err")

				_, _, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		}

		Context("when configured without a blobstore", func() {
			BeforeEach(func() {
				blobs = boshidx.NewFSIndexBlobs("/dir/sub-dir", reporter, nil, sha1calc, fs)
			})

			itCopiesFileIntoDir()
		})

		Context("when configured with a blobstore", func() {
			BeforeEach(func() {
				blobstore = fakeblob.NewFakeBlobstore()
				blobs = boshidx.NewFSIndexBlobs("/dir/sub-dir", reporter, blobstore, sha1calc, fs)
			})

			itCopiesFileIntoDir()

			It("uploads blob and returns blob id", func() {
				blobstore.CreateBlobID = "blob-id"

				blobID, path, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).ToNot(HaveOccurred())
				Expect(blobID).To(Equal("blob-id"))
				Expect(path).To(Equal("/dir/sub-dir/sha1"))

				Expect(blobstore.CreateFileNames).To(Equal([]string{"/tmp/sha1"}))

				Expect(reporter.IndexEntryUploadStartedCallCount()).To(Equal(1))
				Expect(reporter.IndexEntryUploadFinishedCallCount()).To(Equal(1))

				kind, desc := reporter.IndexEntryUploadStartedArgsForCall(0)
				Expect(kind).To(Equal("name"))
				Expect(desc).To(Equal("sha1=sha1"))

				kind, desc, err = reporter.IndexEntryUploadFinishedArgsForCall(0)
				Expect(kind).To(Equal("name"))
				Expect(desc).To(Equal("sha1=sha1"))
				Expect(err).To(BeNil())
			})

			It("returns error if uploading blob fails", func() {
				blobstore.CreateErr = errors.New("fake-err")

				_, _, err := blobs.Add("name", "/tmp/sha1", "sha1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
				Expect(err.Error()).To(ContainSubstring("Creating blob for path '/tmp/sha1'"))

				Expect(reporter.IndexEntryUploadStartedCallCount()).To(Equal(1))
				Expect(reporter.IndexEntryUploadFinishedCallCount()).To(Equal(1))

				kind, desc := reporter.IndexEntryUploadStartedArgsForCall(0)
				Expect(kind).To(Equal("name"))
				Expect(desc).To(Equal("sha1=sha1"))

				kind, desc, err = reporter.IndexEntryUploadFinishedArgsForCall(0)
				Expect(kind).To(Equal("name"))
				Expect(desc).To(Equal("sha1=sha1"))
				Expect(err).ToNot(BeNil())
			})
		})
	})
})
