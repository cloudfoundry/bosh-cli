package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
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
  ami:
    us-east-1: fake-ami-version
    `
		fs.WriteFileString("fake-extracted-path/stemcell.MF", manifestContents)

		applySpecContents := `
		{
		    "packages": {
		        "ruby": {
		            "name": "ruby",
		            "version": "81fe80099130f04dff2b08c80c04e0cb33a9d3ca",
		            "sha1": "94e21ae50bb81200f8b0e50072a1488585e404d3",
		            "blobstore_id": "2226ddfc-0efc-4146-5b41-f979290eadce"
		        },
		        "nats": {
		            "name": "nats",
		            "version": "6a31c7bb0d5ffa2a9f43c7fd7193193438e20e92",
		            "sha1": "2df498116261ad7cee8c6b22dc1da6a2e7e8cc6d",
		            "blobstore_id": "6ab700d0-7499-4d84-62cf-118f8c1a2bad"
		        }
		    },
				"job": {
						"name": "micro_aws",
						"templates": [{
								"name": "nats",
								"version": "b619fdd39f344ee9fb1a760d0ade9ebf9ddcefe2",
								"sha1": "801f569419f160428c147b43f28b10789f93567d",
								"blobstore_id": "c4f81eab-10df-4c42-af88-db0b8269f43f"
						}, {
								"name": "redis",
								"version": "8b447660541a0658b2cbed6f19932c769f037188",
								"sha1": "8738134e8a3352dc1408384bb1883571fa3633f3",
								"blobstore_id": "b62845cd-98de-454a-9d67-4717b24d8412"
						}]
				}
		}`
		fs.WriteFileString("fake-extracted-path/apply_spec.yml", applySpecContents)
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
		expectedStemcell := NewExtractedStemcell(
			Manifest{
				Name:      "fake-stemcell-name",
				Version:   "2690",
				ImagePath: "fake-extracted-path/image",
				CloudProperties: bmproperty.Map{
					"infrastructure": "aws",
					"ami": bmproperty.Map{
						"us-east-1": "fake-ami-version",
					},
				},
			},
			ApplySpec{
				Packages: map[string]Blob{
					"ruby": Blob{
						Name:        "ruby",
						Version:     "81fe80099130f04dff2b08c80c04e0cb33a9d3ca",
						SHA1:        "94e21ae50bb81200f8b0e50072a1488585e404d3",
						BlobstoreID: "2226ddfc-0efc-4146-5b41-f979290eadce",
					},
					"nats": Blob{
						Name:        "nats",
						Version:     "6a31c7bb0d5ffa2a9f43c7fd7193193438e20e92",
						SHA1:        "2df498116261ad7cee8c6b22dc1da6a2e7e8cc6d",
						BlobstoreID: "6ab700d0-7499-4d84-62cf-118f8c1a2bad",
					},
				},
				Job: Job{
					Name: "micro_aws",
					Templates: []Blob{
						{
							Name:        "nats",
							Version:     "b619fdd39f344ee9fb1a760d0ade9ebf9ddcefe2",
							SHA1:        "801f569419f160428c147b43f28b10789f93567d",
							BlobstoreID: "c4f81eab-10df-4c42-af88-db0b8269f43f",
						},
						{
							Name:        "redis",
							Version:     "8b447660541a0658b2cbed6f19932c769f037188",
							SHA1:        "8738134e8a3352dc1408384bb1883571fa3633f3",
							BlobstoreID: "b62845cd-98de-454a-9d67-4717b24d8412",
						},
					},
				},
			},
			"fake-extracted-path",
			fs,
		)
		Expect(stemcell).To(Equal(expectedStemcell))
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
