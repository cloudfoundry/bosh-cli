package index_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshidx "github.com/cloudfoundry/bosh-cli/releasedir/index"
	fakeidx "github.com/cloudfoundry/bosh-cli/releasedir/index/indexfakes"
)

var _ = Describe("FSIndex", func() {
	var (
		reporter *fakeidx.FakeReporter
		blobs    *fakeidx.FakeIndexBlobs
		fs       *fakesys.FakeFileSystem
		index    boshidx.FSIndex
	)

	BeforeEach(func() {
		reporter = &fakeidx.FakeReporter{}
		blobs = &fakeidx.FakeIndexBlobs{}
		fs = fakesys.NewFakeFileSystem()
		index = boshidx.NewFSIndex("index-name", "/dir", true, true, reporter, blobs, fs)
	})

	Describe("Find", func() {
		It("returns nothing if entry with fingerprint is not found", func() {
			path, sha1, err := index.Find("name", "fp")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(BeEmpty())
			Expect(sha1).To(BeEmpty())
		})

		It("returns path and sha1 based on sha1 if entry with fingerprint is found", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  fp2: {version: fp2, sha1: fp2-sha1}
  fp: {version: fp, sha1: fp-sha1}
format-version: "2"`)

			blobs.GetStub = func(name string, blobID string, sha1 string) (string, error) {
				Expect(name).To(Equal("name/fp"))
				Expect(blobID).To(Equal(""))
				Expect(sha1).To(Equal("fp-sha1"))
				return "path", nil
			}

			path, sha1, err := index.Find("name", "fp")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("path"))
			Expect(sha1).To(Equal("fp-sha1"))
		})

		It("returns path and sha1 based on sha1 if entry with fingerprint is found in non-prefixed index file", func() {
			index = boshidx.NewFSIndex("index-name", "/dir", false, true, reporter, blobs, fs)

			fs.WriteFileString("/dir/index.yml", `---
builds:
  fp2: {version: fp2, sha1: fp2-sha1}
  fp: {version: fp, sha1: fp-sha1}
format-version: "2"`)

			blobs.GetReturns("path", nil)

			path, sha1, err := index.Find("name", "fp")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("path"))
			Expect(sha1).To(Equal("fp-sha1"))
		})

		It("returns path and sha1 based on blob id and sha1 if entry with fingerprint is found", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  fp2: {version: fp2, sha1: fp2-sha1, blobstore_id: fp2-blob-id}
  fp: {version: fp, sha1: fp-sha1, blobstore_id: fp-blob-id}
format-version: "2"`)

			blobs.GetStub = func(name string, blobID string, sha1 string) (string, error) {
				Expect(name).To(Equal("name/fp"))
				Expect(blobID).To(Equal("fp-blob-id"))
				Expect(sha1).To(Equal("fp-sha1"))
				return "path", nil
			}

			path, sha1, err := index.Find("name", "fp")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("path"))
			Expect(sha1).To(Equal("fp-sha1"))
		})

		It("returns error if found entry cannot be fetched", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  fp: {version: fp, sha1: fp-sha1, blobstore_id: fp-blob-id}
format-version: "2"`)

			blobs.GetReturns("", errors.New("fake-err"))

			_, _, err := index.Find("name", "fp")
			Expect(err).To(Equal(errors.New("fake-err")))
		})

		It("does not require version to equal entry key", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  fp: {version: other-fp}
format-version: "2"`)

			_, _, err := index.Find("name", "fp")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if name is empty", func() {
			_, _, err := index.Find("", "fp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty name"))
		})

		It("returns error if fingerprint is empty", func() {
			_, _, err := index.Find("name", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty fingerprint"))
		})

		It("returns error if index file cannot be unmarshalled", func() {
			fs.WriteFileString("/dir/name/index.yml", "-")

			_, _, err := index.Find("name", "fp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if reading index file fails", func() {
			fs.WriteFileString("/dir/name/index.yml", "")
			fs.ReadFileError = errors.New("fake-err")

			_, _, err := index.Find("name", "fp")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Add", func() {
		It("adds new entry when no index file exists", func() {
			blobs.AddStub = func(name, path, sha1 string) (string, string, error) {
				Expect(name).To(Equal("name/fp"))
				Expect(path).To(Equal("path"))
				Expect(sha1).To(Equal("sha1"))
				return "blob-id", "blob-path", nil
			}

			path, sha1, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("blob-path"))
			Expect(sha1).To(Equal("sha1"))

			Expect(fs.ReadFileString("/dir/name/index.yml")).To(Equal(`builds:
  fp:
    version: fp
    blobstore_id: blob-id
    sha1: sha1
format-version: "2"
`))
		})

		It("adds new entry to a non-prefixed index file", func() {
			index = boshidx.NewFSIndex("index-name", "/dir", false, true, reporter, blobs, fs)

			blobs.AddReturns("blob-id", "blob-path", nil)

			path, sha1, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("blob-path"))
			Expect(sha1).To(Equal("sha1"))

			Expect(fs.ReadFileString("/dir/index.yml")).To(Equal(`builds:
  fp:
    version: fp
    blobstore_id: blob-id
    sha1: sha1
format-version: "2"
`))
		})

		It("adds new entry with blobstore id to existing index if index allows blobs ids", func() {
			index = boshidx.NewFSIndex("index-name", "/dir", true, true, reporter, blobs, fs)

			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  fp2: {version: fp2, sha1: fp2-sha1, blobstore_id: fp2-blob-id}
format-version: "2"`)

			blobs.AddReturns("blob-id", "blob-path", nil)

			path, sha1, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("blob-path"))
			Expect(sha1).To(Equal("sha1"))

			Expect(fs.ReadFileString("/dir/name/index.yml")).To(Equal(`builds:
  fp:
    version: fp
    blobstore_id: blob-id
    sha1: sha1
  fp2:
    version: fp2
    blobstore_id: fp2-blob-id
    sha1: fp2-sha1
format-version: "2"
`))
		})

		It("adds new entry without blobstore id if index disallows blobs ids", func() {
			index = boshidx.NewFSIndex("index-name", "/dir", true, false, reporter, blobs, fs)

			blobs.AddReturns("", "blob-path", nil)

			path, sha1, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("blob-path"))
			Expect(sha1).To(Equal("sha1"))

			Expect(fs.ReadFileString("/dir/name/index.yml")).To(Equal(`builds:
  fp:
    version: fp
    sha1: sha1
format-version: "2"
`))
		})

		It("returns error when adding entry with blobstore id if index disallows blob ids", func() {
			index = boshidx.NewFSIndex("index-name", "/dir", true, false, reporter, blobs, fs)

			blobs.AddReturns("blob-id", "blob-path", nil)

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				`Internal inconsistency: entry must not include blob ID 'index.indexEntry{Key:"fp", Version:"fp", BlobstoreID:"blob-id", SHA1:"sha1"}'`))
		})

		It("returns error when adding new entry without blobstore id if index allows blob ids", func() {
			index = boshidx.NewFSIndex("index-name", "/dir", true, true, reporter, blobs, fs)

			blobs.AddReturns("", "blob-path", nil)

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				`Internal inconsistency: entry must include blob ID 'index.indexEntry{Key:"fp", Version:"fp", BlobstoreID:"", SHA1:"sha1"}'`))
		})

		It("reports addition", func() {
			blobs.AddReturns("blob-id", "blob-path", nil)

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).ToNot(HaveOccurred())

			Expect(reporter.IndexEntryStartedAddingCallCount()).To(Equal(1))
			Expect(reporter.IndexEntryFinishedAddingCallCount()).To(Equal(1))

			kind, desc := reporter.IndexEntryStartedAddingArgsForCall(0)
			Expect(kind).To(Equal("index-name"))
			Expect(desc).To(Equal("name/fp"))

			kind, desc, err = reporter.IndexEntryFinishedAddingArgsForCall(0)
			Expect(kind).To(Equal("index-name"))
			Expect(desc).To(Equal("name/fp"))
			Expect(err).To(BeNil())
		})

		It("reports addition error if blob cannot be added", func() {
			blobs.AddReturns("", "", errors.New("fake-err"))

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(reporter.IndexEntryStartedAddingCallCount()).To(Equal(1))
			Expect(reporter.IndexEntryFinishedAddingCallCount()).To(Equal(1))

			kind, desc := reporter.IndexEntryStartedAddingArgsForCall(0)
			Expect(kind).To(Equal("index-name"))
			Expect(desc).To(Equal("name/fp"))

			kind, desc, err = reporter.IndexEntryFinishedAddingArgsForCall(0)
			Expect(kind).To(Equal("index-name"))
			Expect(desc).To(Equal("name/fp"))
			Expect(err).ToNot(BeNil())
		})

		It("reports addition error if writing index fails", func() {
			blobs.AddReturns("blob-id", "blob-path", nil)

			fs.WriteFileError = errors.New("fake-err")

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(reporter.IndexEntryStartedAddingCallCount()).To(Equal(1))
			Expect(reporter.IndexEntryFinishedAddingCallCount()).To(Equal(1))

			kind, desc := reporter.IndexEntryStartedAddingArgsForCall(0)
			Expect(kind).To(Equal("index-name"))
			Expect(desc).To(Equal("name/fp"))

			kind, desc, err = reporter.IndexEntryFinishedAddingArgsForCall(0)
			Expect(kind).To(Equal("index-name"))
			Expect(desc).To(Equal("name/fp"))
			Expect(err).ToNot(BeNil())
		})

		It("returns error if there is already an entry with same fingerprint", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  fp: {version: fp}
format-version: "2"`)

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				`Trying to add duplicate index entry 'name/fp' and SHA1 'sha1' (conflicts with 'index.indexEntry{Key:"fp", Version:"fp", BlobstoreID:"", SHA1:""}')`))
		})

		It("returns error if name is empty", func() {
			_, _, err := index.Add("", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty name"))
		})

		It("returns error if fingerprint is empty", func() {
			_, _, err := index.Add("name", "", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty fingerprint"))
		})

		It("returns error if path is empty", func() {
			_, _, err := index.Add("name", "fp", "", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty archive path"))
		})

		It("returns error if sha1 is empty", func() {
			_, _, err := index.Add("name", "fp", "path", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty archive SHA1"))
		})

		It("returns error if index file cannot be unmarshalled", func() {
			fs.WriteFileString("/dir/name/index.yml", "-")

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if reading index file fails", func() {
			fs.WriteFileString("/dir/name/index.yml", "")
			fs.ReadFileError = errors.New("fake-err")

			_, _, err := index.Add("name", "fp", "path", "sha1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
