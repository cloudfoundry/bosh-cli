package blobextract_test

import (
	"errors"
	"os"

	. "github.com/cloudfoundry/bosh-init/installation/blobextract"
	testfakes "github.com/cloudfoundry/bosh-init/testutils/fakes"
	fakeblobstore "github.com/cloudfoundry/bosh-utils/blobstore/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extractor", func() {
	var (
		extractor     Extractor
		blobstore     *fakeblobstore.FakeBlobstore
		targetDir     string
		fakeExtractor *testfakes.FakeMultiResponseExtractor
		logger        boshlog.Logger
		fs            *fakesys.FakeFileSystem

		blobID              string
		blobSHA1            string
		decompressionResult string
		fileName            string
	)

	BeforeEach(func() {
		blobstore = fakeblobstore.NewFakeBlobstore()
		targetDir = "fake-target-dir"
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		blobID = "fake-blob-id"
		blobSHA1 = "fake-sha1"
		fileName = "tarball.tgz"
		blobstore.GetFileName = fileName
		decompressionResult = "decompressed 'tarball.tgz' to directory 'fake-target-dir'"

		extractor = NewExtractor(fs, fakeExtractor, blobstore, logger)
	})

	Describe("Cleanup", func() {
		BeforeEach(func() {
			err := extractor.Extract(blobID, blobSHA1, targetDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("deletes the extracted temp file", func() {
			Expect(fs.FileExists(targetDir)).To(BeTrue())
			err := extractor.Cleanup(blobID, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(targetDir)).To(BeFalse())
		})

		It("deletes the stored blob", func() {
			err := extractor.Cleanup(blobID, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobstore.DeleteBlobID).To(Equal(blobID))
		})
	})

	Describe("Extract", func() {
		Context("when the specified blobID exists in the blobstore", func() {
			It("creates the installed package dir if it does not exist", func() {
				Expect(fs.FileExists(targetDir)).To(BeFalse())
				err := extractor.Extract(blobID, blobSHA1, targetDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(fs.FileExists(targetDir)).To(BeTrue())
			})

			It("decompresses the blob into the target dir", func() {
				err := extractor.Extract(blobID, blobSHA1, targetDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeExtractor.DecompressedFiles()).To(ContainElement(decompressionResult))
			})

			It("cleans up the extracted blob file", func() {
				err := extractor.Extract(blobID, blobSHA1, targetDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(blobstore.CleanUpFileName).To(Equal(fileName))
			})

			Context("when the installed package dir already exists", func() {
				BeforeEach(func() {
					fs.MkdirAll(targetDir, os.ModePerm)
				})

				It("decompresses the blob into the target dir", func() {
					Expect(fs.FileExists(targetDir)).To(BeTrue())
					Expect(fakeExtractor.DecompressedFiles()).ToNot(ContainElement(decompressionResult))

					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).ToNot(HaveOccurred())
					Expect(fs.FileExists(targetDir)).To(BeTrue())
					Expect(fakeExtractor.DecompressedFiles()).To(ContainElement(decompressionResult))
				})

				It("does not re-create the target package dir", func() {
					fs.MkdirAllError = errors.New("fake-error")
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("and decompressing the blob fails", func() {
					It("returns an error and doesn't remove the target dir", func() {
						fakeExtractor.SetDecompressBehavior(
							fileName,
							targetDir,
							errors.New("fake-error"))
						Expect(fs.FileExists(targetDir)).To(BeTrue())
						err := extractor.Extract(blobID, blobSHA1, targetDir)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Decompressing compiled package"))
						Expect(fs.FileExists(targetDir)).To(BeTrue())
					})
				})
			})

			Context("when getting the blob from the blobstore errors", func() {
				BeforeEach(func() {
					blobstore.GetError = errors.New("fake-error")
				})

				It("returns an error", func() {
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Getting object from blobstore"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})
			})

			Context("when creating the target dir fails", func() {
				It("return an error", func() {
					fs.MkdirAllError = errors.New("fake-error")
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Creating target dir"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})

				It("cleans up the blob file", func() {
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).ToNot(HaveOccurred())
					Expect(blobstore.CleanUpFileName).To(Equal(fileName))
				})
			})

			Context("when decompressing the blob fails", func() {
				BeforeEach(func() {
					fakeExtractor.SetDecompressBehavior(
						fileName,
						targetDir,
						errors.New("fake-error"))
				})

				It("returns an error", func() {
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Decompressing compiled package"))
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})

				It("cleans up the target dir if it was created by this extractor", func() {
					Expect(fs.FileExists(targetDir)).To(BeFalse())
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).To(HaveOccurred())
					Expect(fs.FileExists(targetDir)).To(BeFalse())
				})
			})

			Context("when cleaning up the downloaded blob errors", func() {
				BeforeEach(func() {
					blobstore.CleanUpErr = errors.New("fake-error")
				})

				It("does not return the error", func() {
					err := extractor.Extract(blobID, blobSHA1, targetDir)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})

	Describe("ChmodExecutables", func() {
		var (
			binGlob  string
			filePath string
		)

		BeforeEach(func() {
			binGlob = "fake-glob/*"
			filePath = "fake-glob/file"
			fs.SetGlob("fake-glob/*", []string{filePath})
			fs.WriteFileString(filePath, "content")
		})

		It("fetches the files", func() {
			fileMode := fs.GetFileTestStat(filePath).FileMode
			Expect(fileMode).To(Equal(os.FileMode(0)))

			err := extractor.ChmodExecutables(binGlob)
			Expect(err).ToNot(HaveOccurred())

			fileMode = fs.GetFileTestStat(filePath).FileMode
			Expect(fileMode).To(Equal(os.FileMode(0755)))
		})
	})
})
