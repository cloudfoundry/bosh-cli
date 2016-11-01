package blobstore_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/cloudfoundry/bosh-utils/blobstore"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshsysfake "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blob Manager", func() {
	var (
		fs       boshsys.FileSystem
		logger   boshlog.Logger
		basePath string
		blobPath string
		blobId   string
		toWrite  io.Reader
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = boshsys.NewOsFileSystem(logger)
		blobId = "105d33ae-655c-493d-bf9f-1df5cf3ca847"
		basePath = os.TempDir()
		blobPath = filepath.Join(basePath, blobId)
		toWrite = bytes.NewReader([]byte("new data"))
	})

	readFile := func(fileIO boshsys.File) []byte {
		fileStat, _ := fileIO.Stat()
		fileBytes := make([]byte, fileStat.Size())
		fileIO.Read(fileBytes)
		return fileBytes
	}

	It("fetches", func() {
		blobManager := NewBlobManager(fs, basePath)
		fs.WriteFileString(blobPath, "some data")

		readOnlyFile, err, _ := blobManager.Fetch(blobId)
		defer fs.RemoveAll(readOnlyFile.Name())

		Expect(err).ToNot(HaveOccurred())
		fileBytes := readFile(readOnlyFile)

		Expect(string(fileBytes)).To(Equal("some data"))
	})

	It("writes", func() {
		blobManager := NewBlobManager(fs, basePath)
		fs.WriteFileString(blobPath, "some data")
		defer fs.RemoveAll(blobPath)

		err := blobManager.Write(blobId, toWrite)
		Expect(err).ToNot(HaveOccurred())

		contents, err := fs.ReadFileString(blobPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(contents).To(Equal("new data"))
	})

	Context("when it writes", func() {
		BeforeEach(func() {
			basePath = filepath.ToSlash(basePath)
			blobPath = filepath.ToSlash(blobPath)
		})

		It("creates and closes the file", func() {
			fs_ := boshsysfake.NewFakeFileSystem()
			blobManager := NewBlobManager(fs_, basePath)
			err := blobManager.Write(blobId, toWrite)
			Expect(err).ToNot(HaveOccurred())
			fileStats, err := fs_.FindFileStats(blobPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileStats.Open).To(BeFalse())
		})

		It("creates file with correct permissions", func() {
			fs_ := boshsysfake.NewFakeFileSystem()
			blobManager := NewBlobManager(fs_, basePath)
			err := blobManager.Write(blobId, toWrite)
			fileStats, err := fs_.FindFileStats(blobPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileStats.FileMode).To(Equal(os.FileMode(0666)))
			Expect(fileStats.Flags).To(Equal(os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
		})
	})

	Context("GetPath", func() {
		BeforeEach(func() {
			blobId = "smurf-24"
		})

		Describe("when file requested does not exist in blobsPath", func() {
			It("returns an error", func() {
				blobManager := NewBlobManager(fs, basePath)

				_, err := blobManager.GetPath("iblob-id-does-not-exist")

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("blob not found"))
			})
		})

		Describe("when file requested exists in blobsPath", func() {
			It("should return the path of a copy of the requested blob", func() {
				blobManager := NewBlobManager(fs, basePath)

				err := fs.WriteFileString(filepath.Join(basePath, blobId), "smurf-content-hello")
				defer fs.RemoveAll(blobPath)

				Expect(err).To(BeNil())

				filename, err := blobManager.GetPath(blobId)
				Expect(err).To(BeNil())
				Expect(fs.ReadFileString(filename)).To(Equal("smurf-content-hello"))
				Expect(filename).ToNot(Equal(filepath.Join(blobPath, blobId)))
			})
		})
	})

	Context("Delete", func() {
		BeforeEach(func() {
			blobId = "smurf-25"
		})

		Describe("when file to be deleted does not exist in blobsPath", func() {
			It("does not freak out", func() {
				blobManager := NewBlobManager(fs, basePath)

				err := blobManager.Delete("hello-i-am-no-one")

				Expect(err).To(BeNil())
			})
		})

		Describe("when file to be deleted exists in blobsPath", func() {
			It("should delete the blob", func() {
				err := fs.WriteFileString(filepath.Join(basePath, blobId), "smurf-content")
				Expect(err).To(BeNil())
				Expect(fs.FileExists(filepath.Join(basePath, blobId))).To(BeTrue())

				blobManager := NewBlobManager(fs, basePath)
				err = blobManager.Delete(blobId)
				Expect(err).To(BeNil())

				Expect(fs.FileExists(filepath.Join(basePath, blobId))).To(BeFalse())
			})
		})
	})
})
