package resource_test

import (
	"errors"
	"os"
	"strings"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakecrypto "github.com/cloudfoundry/bosh-cli/crypto/fakes"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

var _ = Describe("FingerprinterImpl", func() {
	var (
		sha1calc      *fakecrypto.FakeSha1Calculator
		fs            *fakesys.FakeFileSystem
		fingerprinter FingerprinterImpl
	)

	BeforeEach(func() {
		sha1calc = fakecrypto.NewFakeSha1Calculator()
		fs = fakesys.NewFakeFileSystem()
		fingerprinter = NewFingerprinterImpl(sha1calc, fs)
	})

	It("fingerprints all files", func() {
		files := []File{
			NewFile("/tmp/file2", "/tmp"),
			NewFile("/tmp/file1", "/tmp"),
			NewFile("/tmp/file3", "/tmp"),
			NewFile("/tmp/rel/file4", "/tmp"),
		}

		excludeModeFile := NewFile("/tmp/file5", "/tmp")
		excludeModeFile.ExcludeMode = true
		files = append(files, excludeModeFile)

		basenameFile := NewFile("/tmp/rel/file6", "/tmp")
		basenameFile.UseBasename = true
		files = append(files, basenameFile)

		fs.RegisterOpenFile("/tmp/file1", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeDir},
		})

		fs.RegisterOpenFile("/tmp/file2", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
		})

		fs.RegisterOpenFile("/tmp/file3", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{
				FileType: fakesys.FakeFileTypeFile,
				FileMode: os.FileMode(0111),
			},
		})

		fs.RegisterOpenFile("/tmp/rel/file4", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
		})

		fs.RegisterOpenFile("/tmp/file5", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
		})

		fs.RegisterOpenFile("/tmp/rel/file6", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
		})

		sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
			// file1 directory is not sha1-ed
			"/tmp/file2":     fakecrypto.CalculateInput{Sha1: "file2-sha1"},
			"/tmp/file3":     fakecrypto.CalculateInput{Sha1: "file3-sha1"},
			"/tmp/rel/file4": fakecrypto.CalculateInput{Sha1: "file4-sha1"},
			"/tmp/file5":     fakecrypto.CalculateInput{Sha1: "file5-sha1"},
			"/tmp/rel/file6": fakecrypto.CalculateInput{Sha1: "file6-sha1"},
		})

		chunks := []string{
			"v2",             // version
			"file1", "40755", // dir perms
			"file2", "file2-sha1", "100644", // regular file perms
			"file3", "file3-sha1", "100755", // exec file perms
			"file5", "file5-sha1", // excludes mode
			"rel/file4", "file4-sha1", "100644", // relative file
			"file6", "file6-sha1", "100644", // uses basename
			"chunk1", ",chunk2", // sorted chunks
		}

		sha1calc.CalculateStringInputs = map[string]string{
			strings.Join(chunks, ""): "fp",
		}

		fp, err := fingerprinter.Calculate(files, []string{"chunk2", "chunk1"})
		Expect(err).ToNot(HaveOccurred())
		Expect(fp).To(Equal("fp"))
	})

	It("Includes symlink target in fingerprint calculation", func() {
		files := []File{
			NewFile("/tmp/regular", "/tmp"),
			NewFile("/tmp/symlink", "/tmp"),
		}

		fs.WriteFileString("/tmp/regular", "")
		fs.Symlink("nothing", "/tmp/symlink")

		sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
			"/tmp/regular": fakecrypto.CalculateInput{Sha1: "regular-sha1"},
		})

		chunks := []string{
			"v2", // version
			"regular", "regular-sha1", "100644",
			"symlink", "symlink-target-sha1", "symlink",
			"chunk1", ",chunk2", // sorted chunks
		}

		sha1calc.CalculateStringInputs = map[string]string{
			"nothing":                "symlink-target-sha1",
			strings.Join(chunks, ""): "fp",
		}

		fp, err := fingerprinter.Calculate(files, []string{"chunk2", "chunk1"})
		Expect(err).ToNot(HaveOccurred())
		Expect(fp).To(Equal("fp"))
	})

	It("returns error if stating file fails", func() {
		fs.RegisterOpenFile("/tmp/file2", &fakesys.FakeFile{
			StatErr: errors.New("fake-err"),
		})

		_, err := fingerprinter.Calculate([]File{NewFile("/tmp/file2", "/tmp")}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})

	It("returns error if calculating file sha1 fails", func() {
		fs.RegisterOpenFile("/tmp/file2", &fakesys.FakeFile{
			Stats: &fakesys.FakeFileStats{FileType: fakesys.FakeFileTypeFile},
		})

		sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
			"/tmp/file2": fakecrypto.CalculateInput{Err: errors.New("fake-err")},
		})

		_, err := fingerprinter.Calculate([]File{NewFile("/tmp/file2", "/tmp")}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})
})
