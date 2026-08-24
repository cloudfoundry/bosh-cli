package resource_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/crypto/cryptofakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

var _ = Describe("FingerprinterImpl", func() {
	var (
		digestCalculator *cryptofakes.FakeDigestCalculator
		fs               *fakesys.FakeFileSystem
		fingerprinter    FingerprinterImpl
		followSymlinks   bool
	)

	BeforeEach(func() {
		digestCalculator = &cryptofakes.FakeDigestCalculator{}
		fs = fakesys.NewFakeFileSystem()
	})

	JustBeforeEach(func() {
		fingerprinter = NewFingerprinterImpl(digestCalculator, fs, followSymlinks)
	})

	Context("successfully creating a fingerprint", func() {
		var (
			chunks []string
			files  []File
		)

		BeforeEach(func() {
			files = []File{
				NewFile(filepath.Join("/", "tmp", "file2"), filepath.Join("/", "tmp")),
				NewFile(filepath.Join("/", "tmp", "file1"), filepath.Join("/", "tmp")),
				NewFile(filepath.Join("/", "tmp", "file3"), filepath.Join("/", "tmp")),
				NewFile(filepath.Join("/", "tmp", "rel", "file4"), filepath.Join("/", "tmp")),
			}

			excludeModeFile := NewFile(filepath.Join("/", "tmp", "file5"), filepath.Join("/", "tmp"))
			excludeModeFile.ExcludeMode = true
			files = append(files, excludeModeFile)

			basenameFile := NewFile(filepath.Join("/", "tmp", "rel", "file6"), filepath.Join("/", "tmp"))
			basenameFile.UseBasename = true
			files = append(files, basenameFile)

			fs.RegisterOpenFile(filepath.Join("/", "tmp", "file1"), &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeDir},
			})

			fs.RegisterOpenFile(filepath.Join("/", "tmp", "file2"), &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
			})

			fs.RegisterOpenFile(filepath.Join("/", "tmp", "file3"), &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{
					FileType: fakesys.FakeFileTypeFile,
					FileMode: os.FileMode(0111),
				},
			})

			fs.RegisterOpenFile(filepath.Join("/", "tmp", "rel", "file4"), &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
			})

			fs.RegisterOpenFile(filepath.Join("/", "tmp", "file5"), &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
			})

			fs.RegisterOpenFile(filepath.Join("/", "tmp", "rel", "file6"), &fakesys.FakeFile{
				Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
			})

			digestCalculator.CalculateStub = func(path string) (string, error) {
				switch path {
				// file1 directory is not sha1-ed
				case filepath.Join("/", "tmp", "file1"):
					return "", nil
				case filepath.Join("/", "tmp", "file2"):
					return "file2-sha1", nil
				case filepath.Join("/", "tmp", "file3"):
					return "file3-sha1", nil
				case filepath.Join("/", "tmp", "rel", "file4"):
					return "file4-sha1", nil
				case filepath.Join("/", "tmp", "file5"):
					return "file5-sha1", nil
				case filepath.Join("/", "tmp", "rel", "file6"):
					return "file6-sha1", nil
				default:
					return "", fmt.Errorf("unexpected input '%s'", path)
				}
			}

			chunks = []string{
				"v2",             // version
				"file1", "40755", // dir perms
				"file2", "file2-sha1", "100644", // regular file perms
				"file3", "file3-sha1", "100755", // exec file perms
				"file5", "file5-sha1", // excludes mode
				"rel/file4", "file4-sha1", "100644", // relative file
				"file6", "file6-sha1", "100644", // uses basename
				"chunk1", ",chunk2", // sorted chunks
			}
		})

		It("fingerprints all files", func() {
			digestCalculator.CalculateStringStub = func(input string) string {
				switch input {
				case strings.Join(chunks, ""):
					return "fp"
				default:
					panic(fmt.Sprintf("unexpected input: '%s'", input))
				}
			}

			fp, err := fingerprinter.Calculate(files, []string{"chunk2", "chunk1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fp).To(Equal("fp"))
		})

		It("trims `sha256` algorithm info from resulting fingerprint string", func() {
			digestCalculator.CalculateStringStub = func(input string) string {
				switch input {
				case strings.Join(chunks, ""):
					return "sha256:asdfasdfasdfasdf"
				default:
					panic(fmt.Sprintf("unexpected input: '%s'", input))
				}
			}

			fp, err := fingerprinter.Calculate(files, []string{"chunk2", "chunk1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fp).To(Equal("asdfasdfasdfasdf"))
		})
	})

	It("returns an error when the resulting checksum contains unexpected content so it does not pass incompatible fingerprints to director", func() {
		files := []File{NewFile(filepath.Join("/", "tmp", "file"), filepath.Join("/", "tmp"))}
		err := fs.WriteFileString(filepath.Join("/", "tmp", "file"), "stuff")
		Expect(err).ToNot(HaveOccurred())

		digestCalculator.CalculateStringStub = func(input string) string {
			switch input {
			case strings.Join([]string{"v2", "file", "100644"}, ""):
				return "whatTheAlgorithmIsThat!:asdfasdfasdfasdf"
			default:
				panic(fmt.Sprintf("unexpected input: '%s'", input))
			}
		}

		_, err = fingerprinter.Calculate(files, []string{})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Generated fingerprint contains unexpected characters 'whatTheAlgorithmIsThat!:asdfasdfasdfasdf'"))
	})

	Context("when following symlinks", func() {
		BeforeEach(func() {
			followSymlinks = true
		})

		It("Includes symlink target in fingerprint calculation", func() {
			files := []File{
				NewFile(filepath.Join("/", "tmp", "regular"), filepath.Join("/", "tmp")),
				NewFile(filepath.Join("/", "tmp", "symlink"), filepath.Join("/", "tmp")),
			}

			err := fs.WriteFileString(filepath.Join("/", "tmp", "regular"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.Symlink(filepath.Join("/", "tmp", "regular"), filepath.Join("/", "tmp", "symlink"))
			Expect(err).ToNot(HaveOccurred())

			digestCalculator.CalculateStub = func(path string) (string, error) {
				switch path {
				case filepath.Join("/", "tmp", "regular"):
					return "regular-sha1", nil
				default:
					return "", fmt.Errorf("unexpected input '%s'", path)
				}
			}

			chunks := []string{
				"v2", // version
				"regular", "regular-sha1", "100644",
				"symlink", "regular-sha1", "100644",
				"chunk1", ",chunk2", // sorted chunks
			}

			digestCalculator.CalculateStringStub = func(input string) string {
				switch input {
				case strings.Join(chunks, ""):
					return "fp"
				default:
					panic(fmt.Sprintf("unexpected input: '%s'", input))
				}
			}

			fp, err := fingerprinter.Calculate(files, []string{"chunk2", "chunk1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fp).To(Equal("fp"))
		})
	})

	Context("when not following symlinks", func() {
		BeforeEach(func() {
			followSymlinks = false
		})

		It("Includes symlink target in fingerprint calculation", func() {
			files := []File{
				NewFile(filepath.Join("/", "tmp", "regular"), filepath.Join("/", "tmp")),
				NewFile(filepath.Join("/", "tmp", "symlink"), filepath.Join("/", "tmp")),
			}

			err := fs.WriteFileString(filepath.Join("/", "tmp", "regular"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.Symlink("nothing", filepath.Join("/", "tmp", "symlink"))
			Expect(err).ToNot(HaveOccurred())

			digestCalculator.CalculateStub = func(path string) (string, error) {
				switch path {
				case filepath.Join("/", "tmp", "regular"):
					return "regular-sha1", nil
				default:
					return "", fmt.Errorf("unexpected input '%s'", path)
				}
			}

			chunks := []string{
				"v2", // version
				"regular", "regular-sha1", "100644",
				"symlink", "symlink-target-sha1", "symlink",
				"chunk1", ",chunk2", // sorted chunks
			}

			digestCalculator.CalculateStringStub = func(input string) string {
				switch input {
				case "nothing":
					return "symlink-target-sha1"
				case strings.Join(chunks, ""):
					return "fp"
				default:
					panic(fmt.Sprintf("unexpected input: '%s'", input))
				}
			}

			fp, err := fingerprinter.Calculate(files, []string{"chunk2", "chunk1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fp).To(Equal("fp"))
		})
	})

	It("returns error if stating file fails", func() {
		fs.RegisterOpenFile(filepath.Join("/", "tmp", "file2"), &fakesys.FakeFile{
			StatErr: errors.New("fake-err"),
		})

		_, err := fingerprinter.Calculate([]File{NewFile(filepath.Join("/", "tmp", "file2"), filepath.Join("/", "tmp"))}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})

	It("returns error if calculating file sha1 fails", func() {
		fs.RegisterOpenFile(filepath.Join("/", "tmp", "file2"), &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
		})

		digestCalculator.CalculateStub = func(path string) (string, error) {
			switch path {
			case filepath.Join("/", "tmp", "file2"):
				return "", errors.New("fake-err")
			default:
				return "", fmt.Errorf("unexpected input '%s'", path)
			}
		}

		_, err := fingerprinter.Calculate([]File{NewFile(filepath.Join("/", "tmp", "file2"), filepath.Join("/", "tmp"))}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})
})
