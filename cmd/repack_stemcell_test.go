package cmd_test

import (
	"fmt"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"encoding/json"
	"path/filepath"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	"github.com/cloudfoundry/bosh-cli/stemcell"
	"github.com/cloudfoundry/bosh-cli/stemcell/stemcellfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	"gopkg.in/yaml.v2"
)

var _ = Describe("RepackStemcellCmd", func() {
	var (
		fs      *fakesys.FakeFileSystem
		ui      *fakeui.FakeUI
		command RepackStemcellCmd
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}

	})

	Describe("Run", func() {
		var (
			extractor       stemcell.Extractor
			opts            RepackStemcellOpts
			outputStemcell  string
			compressor      *FakeCompressor
			initialManifest stemcell.Manifest
		)

		BeforeEach(func() {
			compressor = NewFakeCompressor(fs)
			reader := stemcell.NewReader(compressor, fs)
			extractor = stemcell.NewExtractor(reader, fs)
			command = NewRepackStemcellCmd(ui, fs, extractor)
			opts = RepackStemcellOpts{}
		})

		act := func() error { return command.Run(opts) }

		initialManifest = stemcell.Manifest{
			Name:            "name",
			Version:         "1",
			OS:              "fake-os",
			SHA1:            "sha1",
			BoshProtocol:    "1",
			StemcellFormats: []string{"fake-raw"},
			CloudProperties: biproperty.Map{},
		}

		Context("when stemcell path is a local file", func() {
			BeforeEach(func() {
				fs.GlobStub = func(pattern string) ([]string, error) {
					matches := []string{"stemcell.MF", "image"}
					return matches, nil
				}
				scTempDir, err := fs.TempDir("stemcell-files")
				Expect(err).ToNot(HaveOccurred())
				outputStemcell = filepath.Join(scTempDir, "repacked-stemcell.tgz")
				opts.Args.PathToResult = FileArg{ExpandedPath: outputStemcell}

				manifestBytes, err := yaml.Marshal(initialManifest)
				Expect(err).ToNot(HaveOccurred())

				sp := filepath.Join(scTempDir, "stemcell.MF")
				ip := filepath.Join(scTempDir, "image")
				fs.WriteFile(ip, []byte("image-contents"))
				fs.WriteFile(sp, manifestBytes)

				compressedPath, err := compressor.CompressFilesInDir(scTempDir)
				Expect(err).ToNot(HaveOccurred())

				opts.Args.PathToStemcell = compressedPath
			})

			Context("when no flags are passed", func() {
				It("duplicates the stemcell and saves to PathToResult", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					extractDir, _ := fs.TempDir("output-files")
					err = compressor.DecompressFileToDir(outputStemcell, extractDir, boshcmd.CompressorOptions{})
					Expect(err).ToNot(HaveOccurred())

					image, err := fs.ReadFile(filepath.Join(extractDir, "image"))
					Expect(err).ToNot(HaveOccurred())
					Expect(image).To(Equal([]byte("image-contents")))

					repackedManifestFile, err := fs.ReadFile(filepath.Join(extractDir, "stemcell.MF"))
					Expect(err).ToNot(HaveOccurred())

					repackedManifest := stemcell.Manifest{}
					err = yaml.Unmarshal(repackedManifestFile, &repackedManifest)
					Expect(err).ToNot(HaveOccurred())
					Expect(repackedManifest).To(Equal(initialManifest))
				})
			})

			It("overrides fields with the values from passed flags", func() {
				opts = RepackStemcellOpts{
					Args:            opts.Args,
					Name:            "other-name",
					Version:         "2",
					EmptyImage:      true,
					Format:          []string{"fake-rawdisk", "fake-tar"},
					CloudProperties: `{"foo": "bar"}`,
				}
				err := act()
				Expect(err).ToNot(HaveOccurred())

				extractDir, _ := fs.TempDir("output-files")
				err = compressor.DecompressFileToDir(outputStemcell, extractDir, boshcmd.CompressorOptions{})
				Expect(err).ToNot(HaveOccurred())
				repackedManifestFile, err := fs.ReadFile(filepath.Join(extractDir, "stemcell.MF"))
				Expect(err).ToNot(HaveOccurred())

				image, err := fs.ReadFile(filepath.Join(extractDir, "image"))
				Expect(err).ToNot(HaveOccurred())
				Expect(image).To(BeEmpty())

				repackedManifest := stemcell.Manifest{}
				err = yaml.Unmarshal(repackedManifestFile, &repackedManifest)
				Expect(err).ToNot(HaveOccurred())
				Expect(repackedManifest).To(Equal(stemcell.Manifest{
					Name:            "other-name",
					Version:         "2",
					SHA1:            "sha1",
					OS:              "fake-os",
					BoshProtocol:    "1",
					StemcellFormats: []string{"fake-rawdisk", "fake-tar"},
					CloudProperties: biproperty.Map{"foo": "bar"},
				}))
			})

			Context("manifest has no stemcell format", func() {
				initialManifest = stemcell.Manifest{
					Name:            "name",
					Version:         "1",
					OS:              "fake-os",
					SHA1:            "sha1",
					BoshProtocol:    "1",
					CloudProperties: biproperty.Map{},
				}

				It("does not add stemcell formats if they are not present", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					extractDir, _ := fs.TempDir("output-files")
					err = compressor.DecompressFileToDir(outputStemcell, extractDir, boshcmd.CompressorOptions{})
					Expect(err).ToNot(HaveOccurred())

					repackedManifestFile, err := fs.ReadFile(filepath.Join(extractDir, "stemcell.MF"))
					Expect(err).ToNot(HaveOccurred())

					repackedManifest := stemcell.Manifest{}
					err = yaml.Unmarshal(repackedManifestFile, &repackedManifest)
					Expect(err).ToNot(HaveOccurred())
					Expect(repackedManifest).To(Equal(initialManifest))
				})
			})
		})

		Context("when error ocurrs", func() {
			var (
				extractedStemcell *stemcellfakes.FakeExtractedStemcell
				extractor         *stemcellfakes.FakeExtractor
				err               error
			)

			BeforeEach(func() {
				opts = RepackStemcellOpts{}
				opts.Args.PathToStemcell = "some-stemcell.tgz"
				opts.Args.PathToResult = FileArg{ExpandedPath: "repacked-stemcell.tgz"}
				extractor = stemcellfakes.NewFakeExtractor()
				extractedStemcell = &stemcellfakes.FakeExtractedStemcell{}
				command = NewRepackStemcellCmd(ui, fs, extractor)
			})

			Context("and properties are not valid YAML", func() {
				BeforeEach(func() {
					opts.CloudProperties = "not-valid-yaml"
				})

				It("should return an error", func() {
					err = act()
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when it's NOT able to extract stemcell", func() {
				BeforeEach(func() {
					extractor.SetExtractBehavior("some-stemcell.tgz", nil, errors.New("fake-error"))
					err = act()
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when it's NOT able to create new stemcell", func() {
				BeforeEach(func() {
					extractor.SetExtractBehavior("some-stemcell.tgz", extractedStemcell, nil)
					extractedStemcell.PackReturns(errors.New("fake-error"))
					err = act()
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(len(extractor.ExtractInputs)).To(Equal(1))
					Expect(extractor.ExtractInputs[0].TarballPath).To(Equal("some-stemcell.tgz"))

					Expect(extractedStemcell.PackCallCount()).To(Equal(1))
				})
			})
		})
	})
})

type FakeCompressor struct {
	fs *fakesys.FakeFileSystem
}

var _ boshcmd.Compressor = new(FakeCompressor)

func NewFakeCompressor(fs *fakesys.FakeFileSystem) *FakeCompressor {
	return &FakeCompressor{fs: fs}
}

// CompressFilesInDir returns path to a compressed file
func (fc *FakeCompressor) CompressFilesInDir(dir string) (path string, err error) {
	filesInDir, _ := fc.fs.Glob(filepath.Join(dir, "*"))
	archive := map[string][]byte{}

	for _, file := range filesInDir {
		fileContents, err := fc.fs.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return "", fmt.Errorf("Reading file to compress: %s", err.Error())
		}
		archive[file] = fileContents
	}
	compressedArchive, err := json.Marshal(archive)
	if err != nil {
		return "", fmt.Errorf("Marshalling json: %s", err.Error())
	}
	compressedFile, _ := fc.fs.TempFile("stemcell-tgz")
	path = compressedFile.Name()
	fc.fs.WriteFile(compressedFile.Name(), compressedArchive)
	return path, nil
}

func (fc *FakeCompressor) CompressSpecificFilesInDir(dir string, files []string) (path string, err error) {
	return "", errors.New("Not implemented")
}

func (fc *FakeCompressor) DecompressFileToDir(path string, dir string, options boshcmd.CompressorOptions) (err error) {
	archive, err := fc.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Reading file to decompress: %s", err.Error())
	}

	decompressed := map[string][]byte{}
	err = json.Unmarshal(archive, &decompressed)
	if err != nil {
		return fmt.Errorf("Unmarshalling files: %s", err.Error())
	}

	for file, content := range decompressed {
		err = fc.fs.WriteFile(filepath.Join(dir, file), content)
		if err != nil {
			return fmt.Errorf("Writing decompressed files: %s", err.Error())
		}
	}
	return nil
}

// CleanUp cleans up compressed file after it was used
func (fc *FakeCompressor) CleanUp(path string) error {
	return errors.New("Not implemented")
}
