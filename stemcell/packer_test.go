package stemcell_test

import (
	. "github.com/cloudfoundry/bosh-cli/stemcell"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"

	boshcmdfakes "github.com/cloudfoundry/bosh-utils/fileutil/fakes"

	"errors"
	fakebistemcell "github.com/cloudfoundry/bosh-cli/stemcell/stemcellfakes"
	"os"
)

var _ = Describe("Manager", func() {
	var (
		packer                Packer
		fs                    *fakesys.FakeFileSystem
		stemcellExtractionDir string

		fakeCompressor        *boshcmdfakes.FakeCompressor
		fakeExtractedStemcell *fakebistemcell.FakeExtractedStemcell
		tmpDir                string
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		stemcellExtractionDir = "/path/to/dest"
		fs.TempDirDir = stemcellExtractionDir

		fakeExtractedStemcell = new(fakebistemcell.FakeExtractedStemcell)
		fakeCompressor = boshcmdfakes.NewFakeCompressor()
		packer = NewPacker(fakeCompressor)

		fakeExtractedStemcell.GetExtractedPathReturns(stemcellExtractionDir)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Pack", func() {
		Context("when the packaging succeeeds", func() {
			var (
				err         error
				tarballPath string
			)

			BeforeEach(func() {
				fakeCompressor.CompressFilesInDirTarballPath = "a/b/tarball.tgz"
				fakeCompressor.CompressFilesInDirErr = nil

				tarballPath, err = packer.Pack(fakeExtractedStemcell)
			})

			It("packs the extracted stemcell", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.GetExtractedPathCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.SaveCallCount()).To(Equal(1))
				Expect(fakeCompressor.CompressFilesInDirDir).To(Equal(stemcellExtractionDir))

			})

			It("deletes the temporary stemcell file path", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeExtractedStemcell.DeleteCallCount()).To(Equal(1))
			})

			It("returns the path of the created tarball", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.GetExtractedPathCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.SaveCallCount()).To(Equal(1))
				Expect(fakeCompressor.CompressFilesInDirDir).To(Equal(stemcellExtractionDir))

				Expect(tarballPath).To(Equal("a/b/tarball.tgz"))
			})
		})

		Context("when the packaging fails", func() {
			var err error
			Context("when the extracted stemcell can't save its contents", func() {
				BeforeEach(func() {
					fakeExtractedStemcell.SaveReturns(errors.New("fake-error"))
					_, err = packer.Pack(fakeExtractedStemcell)
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())

					Expect(fakeExtractedStemcell.SaveCallCount()).To(Equal(1))
				})
			})

			Context("when the compressor can't create .tgz file", func() {
				BeforeEach(func() {
					fakeExtractedStemcell.SaveReturns(nil)
					fakeCompressor.CompressFilesInDirTarballPath = ""
					fakeCompressor.CompressFilesInDirErr = errors.New("fake-error")
					_, err = packer.Pack(fakeExtractedStemcell)
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())

					Expect(fakeExtractedStemcell.SaveCallCount()).To(Equal(1))
				})

			})

			Context("when the extracted stemcell can't delete its no-longer-needed files", func() {
				BeforeEach(func() {
					fakeExtractedStemcell.SaveReturns(nil)
					fakeCompressor.CompressFilesInDirTarballPath = "fake-path"
					fakeCompressor.CompressFilesInDirErr = nil
					fakeExtractedStemcell.DeleteReturns(errors.New("fake-error"))

					_, err = packer.Pack(fakeExtractedStemcell)
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())

					Expect(fakeExtractedStemcell.SaveCallCount()).To(Equal(1))
					Expect(fakeExtractedStemcell.DeleteCallCount()).To(Equal(1))
				})

			})
		})
	})
})
