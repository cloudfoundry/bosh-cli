package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

var _ = Describe("Reader", func() {
	var (
		compressor     *fakecmd.FakeCompressor
		stemcellReader Reader
		fs             *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		compressor = fakecmd.NewFakeCompressor()
		fs = fakesys.NewFakeFileSystem()
		stemcellReader = NewReader(compressor, fs)

		manifestContents := `
---
name: fake-stemcell-name
version: '2690'
cloud_properties:
  infrastructure: aws
    `
		fs.WriteFileString("fake-extracted-path/stemcell.MF", manifestContents)
	})

	It("extracts the stemcells from a stemcell path", func() {
		_, err := stemcellReader.Read("fake-stemcell-path", "fake-extracted-path")
		Expect(err).ToNot(HaveOccurred())
		Expect(compressor.DecompressFileToDirTarballPaths).To(ContainElement("fake-stemcell-path"))
		Expect(compressor.DecompressFileToDirDirs).To(ContainElement("fake-extracted-path"))
	})

	It("generates correct stemcell", func() {
		stemcell, err := stemcellReader.Read("fake-stemcell-path", "fake-extracted-path")
		Expect(err).ToNot(HaveOccurred())
		Expect(stemcell.Name).To(Equal("fake-stemcell-name"))
		Expect(stemcell.Version).To(Equal("2690"))
		Expect(stemcell.CloudProperties).To(Equal(map[string]interface{}{
			"infrastructure": "aws",
		},
		))

		Expect(stemcell.ImagePath).To(Equal("fake-extracted-path/image"))
	})

	Context("when extracting stemcell fails", func() {
		BeforeEach(func() {
			compressor.DecompressFileToDirErr = errors.New("fake-decompress-error")
		})

		It("returns an error", func() {
			_, err := stemcellReader.Read("fake-stemcell-path", "fake-extracted-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-decompress-error"))
		})
	})

	Context("when reading stemcell manifest fails", func() {
		BeforeEach(func() {
			fs.ReadFileError = errors.New("fake-read-error")
		})

		It("returns an error", func() {
			_, err := stemcellReader.Read("fake-stemcell-path", "fake-extracted-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-read-error"))
		})
	})

	Context("when parsing stemcell manifest fails", func() {
		BeforeEach(func() {
			fs.WriteFileString("fake-extracted-path/stemcell.MF", "")
		})

		It("returns an error", func() {
			_, err := stemcellReader.Read("fake-stemcell-path", "fake-extracted-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Parsing stemcell manifest"))
		})
	})
})
