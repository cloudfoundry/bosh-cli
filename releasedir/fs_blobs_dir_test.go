package releasedir_test

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	fakecrypto "github.com/cloudfoundry/bosh-cli/crypto/fakes"
	fakeblob "github.com/cloudfoundry/bosh-utils/blobstore/fakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/releasedir/releasedirfakes"
)

var _ = Describe("FSBlobsDir", func() {
	var (
		fs        *fakesys.FakeFileSystem
		reporter  *fakereldir.FakeBlobsDirReporter
		blobstore *fakeblob.FakeBlobstore
		sha1calc  *fakecrypto.FakeSha1Calculator
		blobsDir  FSBlobsDir
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		reporter = &fakereldir.FakeBlobsDirReporter{}
		blobstore = fakeblob.NewFakeBlobstore()
		sha1calc = fakecrypto.NewFakeSha1Calculator()
		blobsDir = NewFSBlobsDir("/dir", reporter, blobstore, sha1calc, fs)
	})

	Describe("Blobs", func() {
		act := func() ([]Blob, error) {
			return blobsDir.Blobs()
		}

		It("returns no blobs if blobs.yml is empty", func() {
			fs.WriteFileString("/dir/config/blobs.yml", "")

			blobs, err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(blobs).To(BeEmpty())
		})

		It("returns parsed blobs", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
bosh-116.tgz:
  size: 133959511
  sha: 13ebc5850fcbde216ec32ab4354df53df76e4745
dir/file.tgz:
  size: 133959000
  object_id: ea50bf88-52ca-4230-4ef3-ff22c3975d04
  sha: 2b86b5850fcbde216ec565b4354df53df76e4745
file2.tgz:
  size: 245959511
  object_id: dc21b23e-1e32-40f4-61fb-5c9db26f7375
  sha: 3456b5850fcbde216ec32ab4354df53395607042
`)

			blobs, err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(blobs).To(Equal([]Blob{
				{
					Path: "bosh-116.tgz",
					Size: 133959511,
					SHA1: "13ebc5850fcbde216ec32ab4354df53df76e4745",
				},
				{
					Path:        "dir/file.tgz",
					Size:        133959000,
					BlobstoreID: "ea50bf88-52ca-4230-4ef3-ff22c3975d04",
					SHA1:        "2b86b5850fcbde216ec565b4354df53df76e4745",
				},
				{
					Path:        "file2.tgz",
					Size:        245959511,
					BlobstoreID: "dc21b23e-1e32-40f4-61fb-5c9db26f7375",
					SHA1:        "3456b5850fcbde216ec32ab4354df53395607042",
				},
			}))
		})

		It("returns error if blobs.yml is not found so that user initializes it explicitly", func() {
			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reading blobs index"))
		})

		It("returns error if blobs.yml is not parseable", func() {
			fs.WriteFileString("/dir/config/blobs.yml", "-")

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unmarshalling blobs index"))
		})
	})

	Describe("DownloadBlobs", func() {
		act := func(numOfParallelWorkers int) error {
			return blobsDir.DownloadBlobs(numOfParallelWorkers)
		}

		BeforeEach(func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
dir/file-in-directory.tgz:
  object_id: blob1
  size: 133
  sha: blob1-sha
non-uploaded.tgz:
  size: 245
  sha: 345
file-in-root.tgz:
  object_id: blob2
  size: 245
  sha: blob2-sha
already-downloaded.tgz:
  object_id: blob3
  size: 245
  sha: blob2-sha
`)

			fs.WriteFileString("/blob1-tmp", "blob1-content")
			fs.WriteFileString("/blob2-tmp", "blob2-content")
			fs.WriteFileString("/dir/blobs/already-downloaded.tgz", "blob3-content")

			blobstore.GetFileNames = []string{"/blob1-tmp", "/blob2-tmp"}
		})

		Context("Multiple workers used to download blobs", func() {
			It("downloads all blobs without local blob copy, skipping non-uploaded blobs", func() {
				// Order of blobstore.Get calls may be done in any order
				blobstore := FakeConcurrentBlobstore{
					GetCallback: func(blobID, fingerprint string) (string, error) {
						if blobID == "blob1" && fingerprint == "blob1-sha" {
							return "/blob1-tmp", nil
						} else if blobID == "blob2" && fingerprint == "blob2-sha" {
							return "/blob2-tmp", nil
						} else {
							panic("Received non-matching blobstore.Get call")
						}
					},
				}

				blobsDir = NewFSBlobsDir("/dir", reporter, blobstore, sha1calc, fs)

				err := act(4)
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("/dir/blobs/dir")).To(BeTrue())
				Expect(fs.ReadFileString("/dir/blobs/dir/file-in-directory.tgz")).To(Equal("blob1-content"))
				Expect(fs.ReadFileString("/dir/blobs/file-in-root.tgz")).To(Equal("blob2-content"))
			})
		})

		Context("A single worker to download blobs", func() {
			It("downloads all blobs without local blob copy, skipping non-uploaded blobs", func() {
				err := act(1)
				Expect(err).ToNot(HaveOccurred())

				Expect(blobstore.GetBlobIDs).To(Equal([]string{"blob1", "blob2"}))
				Expect(blobstore.GetFingerprints).To(Equal([]string{"blob1-sha", "blob2-sha"}))

				Expect(fs.FileExists("/dir/blobs/dir")).To(BeTrue())
				Expect(fs.ReadFileString("/dir/blobs/dir/file-in-directory.tgz")).To(Equal("blob1-content"))
				Expect(fs.ReadFileString("/dir/blobs/file-in-root.tgz")).To(Equal("blob2-content"))
			})

			It("reports downloaded blobs skipping already existing ones", func() {
				err := act(1)
				Expect(err).ToNot(HaveOccurred())

				{
					Expect(reporter.BlobDownloadStartedCallCount()).To(Equal(2))

					path, size, blobID, sha1 := reporter.BlobDownloadStartedArgsForCall(0)
					Expect(path).To(Equal("dir/file-in-directory.tgz"))
					Expect(size).To(Equal(int64(133)))
					Expect(blobID).To(Equal("blob1"))
					Expect(sha1).To(Equal("blob1-sha"))

					path, size, blobID, sha1 = reporter.BlobDownloadStartedArgsForCall(1)
					Expect(path).To(Equal("file-in-root.tgz"))
					Expect(size).To(Equal(int64(245)))
					Expect(blobID).To(Equal("blob2"))
					Expect(sha1).To(Equal("blob2-sha"))
				}

				{
					Expect(reporter.BlobDownloadFinishedCallCount()).To(Equal(2))

					path, blobID, err := reporter.BlobDownloadFinishedArgsForCall(0)
					Expect(path).To(Equal("dir/file-in-directory.tgz"))
					Expect(blobID).To(Equal("blob1"))
					Expect(err).ToNot(HaveOccurred())

					path, blobID, err = reporter.BlobDownloadFinishedArgsForCall(1)
					Expect(path).To(Equal("file-in-root.tgz"))
					Expect(blobID).To(Equal("blob2"))
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})

		Context("downloading fails", func() {
			It("reports error", func() {
				blobstore.GetErrs = []error{errors.New("fake-err")}

				err := act(1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting blob 'blob1' for path 'dir/file-in-directory.tgz': fake-err"))

				Expect(reporter.BlobDownloadStartedCallCount()).To(Equal(2))
				Expect(reporter.BlobDownloadFinishedCallCount()).To(Equal(2))
			})

			Context("when more than one blob fails to download", func() {
				It("reports error", func() {
					blobstore.GetErrs = []error{errors.New("fake-err1"), errors.New("fake-err2")}

					err := act(1)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Getting blob 'blob1' for path 'dir/file-in-directory.tgz': fake-err1"))
					Expect(err.Error()).To(ContainSubstring("Getting blob 'blob2' for path 'file-in-root.tgz': fake-err2"))

					Expect(fs.FileExists("/dir/blobs/dir")).To(BeFalse())
					Expect(fs.FileExists("/dir/blobs/dir/file-in-directory.tgz")).To(BeFalse())
					Expect(fs.FileExists("/dir/blobs/file-in-root.tgz")).To(BeFalse())

				})
			})

			Context("without creating any blob sub-dirs", func() {
				It("returns error", func() {
					blobstore.GetErrs = []error{errors.New("fake-err"), errors.New("fake-err")}

					err := act(1)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(fs.FileExists("/dir/blobs/dir")).To(BeFalse())
					Expect(fs.FileExists("/dir/blobs/dir/file-in-directory.tgz")).To(BeFalse())
					Expect(fs.FileExists("/dir/blobs/file-in-root.tgz")).To(BeFalse())
				})
			})

			Context("without placing any local blobs", func() {
				It("returns error", func() {
					blobstore.GetErrs = []error{nil, errors.New("fake-err")}

					err := act(1)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))

					Expect(fs.FileExists("/dir/blobs/dir")).To(BeTrue())
					Expect(fs.FileExists("/dir/blobs/dir/file-in-directory.tgz")).To(BeTrue())
					Expect(fs.FileExists("/dir/blobs/file-in-root.tgz")).To(BeFalse())
				})
			})
		})

		Context("when creating blob sub-dir fails", func() {
			It("returns error", func() {
				fs.MkdirAllError = errors.New("fake-err")

				err := act(1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when moving temp blob file across devices into its final destination", func() {
			BeforeEach(func() {
				fs.RenameError = &os.LinkError{
					Err: syscall.Errno(0x12),
				}
			})

			It("downloads all blobs without local blob copy", func() {
				err := act(1)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when copying blobs across devices fails", func() {
				It("returns error", func() {
					fs.CopyFileError = errors.New("failed to copy")

					err := act(1)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to copy"))
				})
			})
		})

		Context("when moving temp blob file into its final destination fails for an uncaught reason", func() {
			It("returns error", func() {
				fs.RenameError = errors.New("fake-err")

				err := act(1)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})

	Describe("TrackBlob", func() {
		act := func() (Blob, error) {
			content := ioutil.NopCloser(strings.NewReader(string("content")))
			return blobsDir.TrackBlob("dir/file.tgz", content)
		}

		BeforeEach(func() {
			fs.WriteFileString("/dir/config/blobs.yml", "")

			fs.ReturnTempFile = fakesys.NewFakeFile("/tmp-file", fs)

			sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
				"/tmp-file": fakecrypto.CalculateInput{Sha1: "content-sha1"},
			})
		})

		It("adds a blob to the list if it's not already tracked", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
file2.tgz:
  size: 245
  sha: 345
`)

			blob, err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(blob).To(Equal(Blob{Path: "dir/file.tgz", Size: 7, SHA1: "content-sha1"}))

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "dir/file.tgz", Size: 7, SHA1: "content-sha1"},
				{Path: "file2.tgz", Size: 245, SHA1: "345"},
			}))

			Expect(fs.ReadFileString("/dir/blobs/dir/file.tgz")).To(Equal("content"))
		})

		It("updates blob record if it's already tracked", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
dir/file.tgz:
  size: 133
  sha: 13e
file2.tgz:
  size: 245
  sha: 345
`)

			blob, err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(blob).To(Equal(Blob{Path: "dir/file.tgz", Size: 7, SHA1: "content-sha1"}))

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "dir/file.tgz", Size: 7, SHA1: "content-sha1"},
				{Path: "file2.tgz", Size: 245, SHA1: "345"},
			}))

			Expect(fs.ReadFileString("/dir/blobs/dir/file.tgz")).To(Equal("content"))
		})

		It("overrides existing local blob copy", func() {
			fs.WriteFileString("/dir/blobs/dir/file.tgz", "prev-content")

			_, err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/blobs/dir/file.tgz")).To(Equal("content"))
		})

		It("returns error and does not update blobs.yml if temp file cannot be opened", func() {
			fs.TempFileError = errors.New("fake-err")

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(BeEmpty())
		})

		It("returns error and does not update blobs.yml if copying from src fails", func() {
			file := fakesys.NewFakeFile("/tmp-file", fs)
			file.WriteErr = errors.New("fake-err")
			fs.ReturnTempFile = file

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(BeEmpty())
		})

		It("returns error and does not update blobs.yml if cannot determine size", func() {
			file := fakesys.NewFakeFile("/tmp-file", fs)
			file.StatErr = errors.New("fake-err")
			fs.ReturnTempFile = file

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(BeEmpty())
		})

		It("returns error and does not update blobs.yml if calculating sha1 fails", func() {
			sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
				"/tmp-file": fakecrypto.CalculateInput{Err: errors.New("fake-err")},
			})

			_, err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(BeEmpty())
		})
	})

	Describe("UntrackBlob", func() {
		act := func() error {
			return blobsDir.UntrackBlob("dir/file.tgz")
		}

		It("removes reference from list of blobs (first)", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
dir/file.tgz:
  size: 133
  sha: 13e
file2.tgz:
  size: 245
  sha: 345
`)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "file2.tgz", Size: 245, SHA1: "345"},
			}))
		})

		It("removes reference from list of blobs (middle)", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
bosh-116.tgz:
  size: 133
  sha: 13e
dir/file.tgz:
  size: 133
  sha: 2b8
file2.tgz:
  size: 245
  sha: 345
`)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "bosh-116.tgz", Size: 133, SHA1: "13e"},
				{Path: "file2.tgz", Size: 245, SHA1: "345"},
			}))
		})

		It("removes reference from list of blobs (last)", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
bosh-116.tgz:
  size: 133
  sha: 13e
dir/file.tgz:
  size: 245
  sha: 345
`)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "bosh-116.tgz", Size: 133, SHA1: "13e"},
			}))
		})

		It("succeeds even if record is not found", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
bosh-116.tgz:
  size: 133
  sha: 13e
`)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "bosh-116.tgz", Size: 133, SHA1: "13e"},
			}))
		})

		It("removes local blob copy", func() {
			fs.WriteFileString("/dir/config/blobs.yml", "")
			fs.WriteFileString("/dir/blobs/dir/file.tgz", "blob")

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/dir/blobs/dir/file.tgz")).To(BeFalse())
		})

		It("returns error if removing local blob copy fails", func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
dir/file.tgz:
  size: 133
  sha: 13e
`)

			fs.RemoveAllStub = func(_ string) error {
				return errors.New("fake-err")
			}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "dir/file.tgz", Size: 133, SHA1: "13e"},
			}))
		})
	})

	Describe("UploadBlobs", func() {
		act := func() error {
			return blobsDir.UploadBlobs()
		}

		BeforeEach(func() {
			fs.WriteFileString("/dir/config/blobs.yml", `
dir/file-in-directory.tgz:
  object_id: blob1
  size: 133
  sha: blob1-sha
non-uploaded.tgz:
  size: 243
  sha: blob2-sha
file-in-root.tgz:
  object_id: blob3
  size: 245
  sha: blob3-sha
already-downloaded.tgz:
  object_id: blob4
  size: 245
  sha: blob4-sha
non-uploaded2.tgz:
  size: 245
  sha: blob5-sha
`)

			blobstore.CreateBlobIDs = []string{"blob2", "blob5"}
		})

		It("uploads non-uploaded blobs", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobstore.CreateFileNames).To(Equal(
				[]string{"/dir/blobs/non-uploaded.tgz", "/dir/blobs/non-uploaded2.tgz"}))

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "already-downloaded.tgz", Size: 245, BlobstoreID: "blob4", SHA1: "blob4-sha"},
				{Path: "dir/file-in-directory.tgz", Size: 133, BlobstoreID: "blob1", SHA1: "blob1-sha"},
				{Path: "file-in-root.tgz", Size: 245, BlobstoreID: "blob3", SHA1: "blob3-sha"},
				{Path: "non-uploaded.tgz", Size: 243, BlobstoreID: "blob2", SHA1: "blob2-sha"},
				{Path: "non-uploaded2.tgz", Size: 245, BlobstoreID: "blob5", SHA1: "blob5-sha"},
			}))
		})

		It("reports uploaded blobs skipping already existing ones", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			{
				Expect(reporter.BlobUploadStartedCallCount()).To(Equal(2))

				path, size, sha1 := reporter.BlobUploadStartedArgsForCall(0)
				Expect(path).To(Equal("non-uploaded.tgz"))
				Expect(size).To(Equal(int64(243)))
				Expect(sha1).To(Equal("blob2-sha"))

				path, size, sha1 = reporter.BlobUploadStartedArgsForCall(1)
				Expect(path).To(Equal("non-uploaded2.tgz"))
				Expect(size).To(Equal(int64(245)))
				Expect(sha1).To(Equal("blob5-sha"))
			}

			{
				Expect(reporter.BlobUploadFinishedCallCount()).To(Equal(2))

				path, blobID, err := reporter.BlobUploadFinishedArgsForCall(0)
				Expect(path).To(Equal("non-uploaded.tgz"))
				Expect(blobID).To(Equal("blob2"))
				Expect(err).ToNot(HaveOccurred())

				path, blobID, err = reporter.BlobUploadFinishedArgsForCall(1)
				Expect(path).To(Equal("non-uploaded2.tgz"))
				Expect(blobID).To(Equal("blob5"))
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("returns error if uploading fails and does not change blobs.yml", func() {
			blobstore.CreateErrs = []error{errors.New("fake-err")}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "already-downloaded.tgz", Size: 245, BlobstoreID: "blob4", SHA1: "blob4-sha"},
				{Path: "dir/file-in-directory.tgz", Size: 133, BlobstoreID: "blob1", SHA1: "blob1-sha"},
				{Path: "file-in-root.tgz", Size: 245, BlobstoreID: "blob3", SHA1: "blob3-sha"},
				{Path: "non-uploaded.tgz", Size: 243, SHA1: "blob2-sha"},
				{Path: "non-uploaded2.tgz", Size: 245, SHA1: "blob5-sha"},
			}))
		})

		It("reports error if uploading fails", func() {
			blobstore.CreateErrs = []error{errors.New("fake-err")}
			blobstore.CreateBlobIDs = []string{}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(reporter.BlobUploadStartedCallCount()).To(Equal(1))
			Expect(reporter.BlobUploadFinishedCallCount()).To(Equal(1))

			path, size, sha1 := reporter.BlobUploadStartedArgsForCall(0)
			Expect(path).To(Equal("non-uploaded.tgz"))
			Expect(size).To(Equal(int64(243)))
			Expect(sha1).To(Equal("blob2-sha"))

			path, blobID, err := reporter.BlobUploadFinishedArgsForCall(0)
			Expect(path).To(Equal("non-uploaded.tgz"))
			Expect(blobID).To(Equal(""))
			Expect(err).To(HaveOccurred())
		})

		It("returns if saving blobstore id fails and does not continue to upload other blobs", func() {
			fs.WriteFileError = errors.New("fake-err")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			// Include blobstore id in error message for cleanup purposes
			Expect(err.Error()).To(ContainSubstring("Saving newly created blob 'blob2'"))

			Expect(reporter.BlobUploadStartedCallCount()).To(Equal(1))
		})

		It("returns error if uploading fails and saves blob id for successfully uploaded blobs", func() {
			blobstore.CreateErrs = []error{nil, errors.New("fake-err")}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.Blobs()).To(Equal([]Blob{
				{Path: "already-downloaded.tgz", Size: 245, BlobstoreID: "blob4", SHA1: "blob4-sha"},
				{Path: "dir/file-in-directory.tgz", Size: 133, BlobstoreID: "blob1", SHA1: "blob1-sha"},
				{Path: "file-in-root.tgz", Size: 245, BlobstoreID: "blob3", SHA1: "blob3-sha"},
				{Path: "non-uploaded.tgz", Size: 243, BlobstoreID: "blob2", SHA1: "blob2-sha"},
				{Path: "non-uploaded2.tgz", Size: 245, SHA1: "blob5-sha"},
			}))
		})
	})
})

type FakeConcurrentBlobstore struct {
	GetCallback func(blobID, fingerprint string) (string, error)
}

func (bs FakeConcurrentBlobstore) Get(blobID, fingerprint string) (string, error) {
	return bs.GetCallback(blobID, fingerprint)
}

func (bs FakeConcurrentBlobstore) CleanUp(fileName string) error                  { return nil }
func (bs FakeConcurrentBlobstore) Delete(blobId string) error                     { return nil }
func (bs FakeConcurrentBlobstore) Create(fileName string) (string, string, error) { return "", "", nil }
func (bs FakeConcurrentBlobstore) Validate() error                                { return nil }
