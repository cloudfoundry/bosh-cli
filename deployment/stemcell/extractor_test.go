package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

var _ = Describe("Manager", func() {
	var (
		repo                *fakebmconfig.FakeStemcellRepo
		extractor           Extractor
		fs                  *fakesys.FakeFileSystem
		reader              *fakebmstemcell.FakeStemcellReader
		stemcellTarballPath string
		tempExtractionDir   string

		expectedExtractedStemcell ExtractedStemcell
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		reader = fakebmstemcell.NewFakeReader()
		repo = fakebmconfig.NewFakeStemcellRepo()
		stemcellTarballPath = "/stemcell/tarball/path"
		tempExtractionDir = "/path/to/dest"
		fs.TempDirDir = tempExtractionDir

		extractor = NewExtractor(reader, fs)

		expectedExtractedStemcell = NewExtractedStemcell(
			Manifest{
				Name:      "fake-stemcell-name",
				ImagePath: "fake-image-path",
				RawCloudProperties: map[interface{}]interface{}{
					"fake-prop-key": "fake-prop-value",
				},
			},
			ApplySpec{},
			tempExtractionDir,
			fs,
		)
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedExtractedStemcell, nil)
	})

	Describe("Extract", func() {
		It("extracts and parses the stemcell manifest", func() {
			stemcell, err := extractor.Extract(stemcellTarballPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).To(Equal(expectedExtractedStemcell))

			Expect(reader.ReadInputs).To(Equal([]fakebmstemcell.ReadInput{
				{
					StemcellTarballPath: stemcellTarballPath,
					DestPath:            tempExtractionDir,
				},
			}))
		})

		It("when the read fails, returns an error", func() {
			reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedExtractedStemcell, errors.New("fake-read-error"))

			_, err := extractor.Extract(stemcellTarballPath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-read-error"))
		})
	})
})
