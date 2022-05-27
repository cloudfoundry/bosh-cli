package stemcell_test

import (
	. "github.com/cloudfoundry/bosh-cli/v6/stemcell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"os"

	"fmt"

	boshcmdfakes "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Stemcell", func() {
	var (
		stemcell      ExtractedStemcell
		manifest      Manifest
		fakefs        *fakesys.FakeFileSystem
		extractedPath string
		compressor    *boshcmdfakes.FakeCompressor
	)

	BeforeEach(func() {
		manifest = Manifest{
			Name: "new-name",
		}

		extractedPath = "fake-path"

		fakefs = fakesys.NewFakeFileSystem()
		compressor = new(boshcmdfakes.FakeCompressor)
		stemcell = NewExtractedStemcell(
			manifest,
			extractedPath,
			compressor,
			fakefs,
		)
	})

	Describe("Manifest", func() {
		It("returns the manifest", func() {
			Expect(stemcell.Manifest()).To(Equal(manifest))
		})

		Context("when stemcell.MF contains api_version", func() {
			BeforeEach(func() {
				manifest = Manifest{
					Name:       "new-name",
					ApiVersion: 42,
				}

				stemcell = NewExtractedStemcell(
					manifest,
					extractedPath,
					compressor,
					fakefs,
				)
			})

			It("it populates api_version in manifest", func() {
				Expect(stemcell.Manifest()).To(Equal(manifest))
			})
		})
	})

	Describe("Delete", func() {
		var removeAllCalled bool

		BeforeEach(func() {
			removeAllCalled = false
			fakefs.RemoveAllStub = func(_ string) error {
				removeAllCalled = true
				return nil
			}
		})

		It("removes the extracted path", func() {
			Expect(stemcell.Cleanup()).To(BeNil())
			Expect(removeAllCalled).To(BeTrue())
		})
	})

	Describe("String", func() {
		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)
		})

		It("returns the name and the version", func() {
			Expect(stemcell.String()).To(Equal("ExtractedStemcell{name=some-name version=some-version}"))
		})
	})

	Describe("OsAndVersion", func() {
		BeforeEach(func() {
			manifest = Manifest{
				OS:      "some-os",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)
		})

		It("returns the name and the version", func() {
			Expect(stemcell.OsAndVersion()).To(Equal("some-os/some-version"))
		})
	})

	Describe("SetName", func() {
		var newStemcellName string

		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)

			newStemcellName = "some-new-name"
		})

		It("sets the name", func() {
			stemcell.SetName(newStemcellName)
			Expect(stemcell.Manifest().Name).To(Equal(newStemcellName))
		})
	})

	Describe("SetVersion", func() {
		var newStemcellVersion string

		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)

			newStemcellVersion = "some-new-version"
		})

		It("sets the name", func() {
			stemcell.SetVersion(newStemcellVersion)
			Expect(stemcell.Manifest().Version).To(Equal(newStemcellVersion))
		})
	})

	Describe("SetCloudProperties", func() {
		var newStemcellCloudProperties biproperty.Map

		BeforeEach(func() {
			manifest = Manifest{CloudProperties: biproperty.Map{}}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)

			newStemcellCloudProperties = biproperty.Map{
				"some_property": "some_value",
			}
		})

		It("sets the properties", func() {
			stemcell.SetCloudProperties(newStemcellCloudProperties)
			Expect(stemcell.Manifest().CloudProperties["some_property"]).To(Equal("some_value"))
		})

		Context("there are existing properties in the MF file", func() {
			BeforeEach(func() {

				cloudProperties := biproperty.Map{
					"some_property":     "to be overwritten",
					"existing_property": "this should stick around",
				}

				manifest = Manifest{CloudProperties: cloudProperties}

				stemcell = NewExtractedStemcell(
					manifest,
					extractedPath,
					compressor,
					fakefs,
				)

				newStemcellCloudProperties = biproperty.Map{
					"some_property": "totally overwritten, dude",
					"new_property":  "didn't previously exist",
				}
			})

			It("overwrites existing properties", func() {
				stemcell.SetCloudProperties(newStemcellCloudProperties)
				Expect(stemcell.Manifest().CloudProperties["some_property"]).To(Equal("totally overwritten, dude"))
			})

			It("does not remove existing properties", func() {
				stemcell.SetCloudProperties(newStemcellCloudProperties)
				Expect(stemcell.Manifest().CloudProperties["existing_property"]).To(Equal("this should stick around"))
			})

			It("adds new properties", func() {
				stemcell.SetCloudProperties(newStemcellCloudProperties)
				Expect(stemcell.Manifest().CloudProperties["new_property"]).To(Equal("didn't previously exist"))
			})
		})
	})

	Describe("EmptyImage", func() {
		var (
			imagePath string
		)

		BeforeEach(func() {
			extractedPath = "extracted-path"
			imagePath = "extracted-path/image"

			fakefs.MkdirAll(extractedPath, os.ModeDir)
			file := fakesys.NewFakeFile(imagePath, fakefs)
			file.Write([]byte("tar-gz-header-and-content"))

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)
		})

		Context("when overwriting the image succeeds", func() {
			It("overwrites the image", func() {
				err := stemcell.EmptyImage()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakefs.WriteFileCallCount).To(Equal(1))
				stat, err := fakefs.Stat(imagePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(stat.Size()).To(Equal(int64(0)))
			})
		})

		Context("when overwriting fails", func() {
			It("returns an error", func() {
				fakefs.WriteFileError = errors.New("could not write file")

				err := stemcell.EmptyImage()
				Expect(err).To(HaveOccurred())
			})
		})

	})

	Describe("SetFormat", func() {
		var newStemcellFormat []string

		BeforeEach(func() {
			manifest = Manifest{
				Name:            "some-name",
				Version:         "some-version",
				StemcellFormats: []string{"some-format"},
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)

			newStemcellFormat = []string{"some-new-format"}
		})

		It("sets the format", func() {
			stemcell.SetFormat(newStemcellFormat)
			Expect(stemcell.Manifest().StemcellFormats).To(Equal(newStemcellFormat))
		})

		Context(" when multiple --format options are specified", func() {
			BeforeEach(func() {
				newStemcellFormat = []string{"some-new-format", "second-new-format"}
			})

			It("overrides the stemcell_formats with all provided formats", func() {
				stemcell.SetFormat(newStemcellFormat)
				Expect(stemcell.Manifest().StemcellFormats).To(Equal(newStemcellFormat))
			})
		})
	})

	Describe("Pack", func() {
		var (
			removeAllCalled bool
			destinationPath string
		)

		BeforeEach(func() {
			removeAllCalled = false
			extractedPath = "extracted-path"
			destinationPath = "destination/tarball.tgz"

			fakefs.MkdirAll("destination", os.ModeDir)

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				compressor,
				fakefs,
			)
		})

		Context("when the packaging succeeeds", func() {
			var compressedTarballPath string

			BeforeEach(func() {
				compressedTarballPath = "bosh-platform-disk-TarballCompressor-CompressSpecificFilesInDir/generated-tarball.tgz"
			})

			DescribeTable("pack stemcell",
				func(apiVersion int) {
					if apiVersion > 0 {
						manifest.ApiVersion = apiVersion
						stemcell = NewExtractedStemcell(
							manifest,
							extractedPath,
							compressor,
							fakefs,
						)
					}
					compressor.CompressFilesInDirTarballPath = compressedTarballPath
					compressor.CompressFilesInDirErr = nil
					compressor.CompressFilesInDirCallBack = func() {
						fakefs.WriteFileString(compressedTarballPath, "hello")
					}

					removeAllCalled = false
					fakefs.RenameError = nil

					fakefs.RemoveAllStub = func(path string) error {
						removeAllCalled = true
						Expect(path).To(Equal(extractedPath))
						// We are returning an error to disable the removal of the directory containing the extracted files,
						// particularly stemcell.MF, which we need to examine to test that OS/Version/Cloud Properties
						// were properly updated.
						return errors.New("Not error.")
					}

					err := stemcell.Pack(destinationPath)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakefs.RenameOldPaths[0]).To(Equal(compressedTarballPath))
					Expect(fakefs.RenameNewPaths[0]).To(Equal("destination/tarball.tgz"))

					Expect(compressor.CompressFilesInDirDir).To(Equal(extractedPath))

					newStemcellMFContent, err := fakefs.ReadFileString("extracted-path/stemcell.MF")
					Expect(err).ToNot(HaveOccurred())
					Expect(newStemcellMFContent).To(ContainSubstring("name: new-name"))
					Expect(newStemcellMFContent).NotTo(ContainSubstring("stemcell_formats:"))
					if apiVersion > 0 {
						Expect(newStemcellMFContent).To(ContainSubstring(fmt.Sprintf("api_version: %d", apiVersion)))
					} else {
						Expect(newStemcellMFContent).NotTo(ContainSubstring("api_version:"))
					}

					Expect(fakefs.FileExists(compressedTarballPath)).To(BeFalse())
					Expect(fakefs.FileExists(destinationPath)).To(BeTrue())

					Expect(removeAllCalled).To(BeTrue())
				},
				Entry("api_version spefied", 42),
				Entry("api_version NOT specified", 0),
			)
		})

		Context("when the packaging fails", func() {
			Context("when the extracted stemcell can't save its contents", func() {
				It("returns an error", func() {
					fakefs.RemoveAllStub = func(path string) error {
						removeAllCalled = true
						Expect(path).To(Equal(extractedPath))
						return errors.New("Not error.")
					}
					fakefs.WriteFileError = errors.New("could not write file")

					err := stemcell.Pack(destinationPath)
					Expect(err).To(HaveOccurred())

					Expect(removeAllCalled).To(BeTrue())
				})
			})

			Context("when the compressor can't create .tgz file", func() {
				It("returns an error", func() {
					compressor.CompressFilesInDirTarballPath = ""
					compressor.CompressFilesInDirErr = errors.New("error creating .tgz file")
					removeAllCalled = false
					fakefs.RemoveAllStub = func(path string) error {
						removeAllCalled = true
						Expect(path).To(Equal(extractedPath))
						return errors.New("Not error.")
					}

					err := stemcell.Pack(destinationPath)
					Expect(err).To(HaveOccurred())

					Expect(compressor.CompressFilesInDirDir).To(Equal(extractedPath))
				})
			})
		})

		Context("when moving the newly-packed stemcell into place fails", func() {
			BeforeEach(func() {
				fakefs.RenameError = errors.New("could not copy file into place")
			})

			It("returns an error", func() {
				err := stemcell.Pack(destinationPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not copy file into place"))

				Expect(removeAllCalled).To(BeFalse())
			})
		})
	})
})
