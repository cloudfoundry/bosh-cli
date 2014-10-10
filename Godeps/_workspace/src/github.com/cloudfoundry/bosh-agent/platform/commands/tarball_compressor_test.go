package commands_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	. "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
)

func fixtureSrcDir() string {
	pwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return filepath.Join(pwd, "..", "..", "Fixtures", "test_filtered_copy_to_temp")
}

func fixtureSrcTgz() string {
	pwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return filepath.Join(pwd, "..", "..", "Fixtures", "compressor-decompress-file-to-dir.tgz")
}

func BeDir() beDirMatcher {
	return beDirMatcher{}
}

type beDirMatcher struct {
}

//FailureMessage(actual interface{}) (message string)
//NegatedFailureMessage(actual interface{}) (message string)
func (m beDirMatcher) Match(actual interface{}) (bool, error) {
	path, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("`%s' is not a valid path", actual)
	}

	dir, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("Could not open `%s'", actual)
	}
	defer dir.Close()

	dirInfo, err := dir.Stat()
	if err != nil {
		return false, fmt.Errorf("Could not stat `%s'", actual)
	}

	return dirInfo.IsDir(), nil
}

func (m beDirMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected `%s' to be a directory", actual)
}

func (m beDirMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected `%s' to not be a directory", actual)
}

var _ = Describe("tarballCompressor", func() {
	var (
		dstDir     string
		cmdRunner  boshsys.CmdRunner
		fs         boshsys.FileSystem
		compressor Compressor
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cmdRunner = boshsys.NewExecCmdRunner(logger)
		fs = boshsys.NewOsFileSystem(logger)
		tmpDir, err := fs.TempDir("tarballCompressor-test")
		Expect(err).NotTo(HaveOccurred())
		dstDir = filepath.Join(tmpDir, "TestCompressor")
		compressor = NewTarballCompressor(cmdRunner, fs)
	})

	BeforeEach(func() {
		fs.MkdirAll(dstDir, os.ModePerm)
	})

	AfterEach(func() {
		fs.RemoveAll(dstDir)
	})

	Describe("CompressFilesInDir", func() {
		It("compresses the files in the given directory", func() {
			srcDir := fixtureSrcDir()
			tgzName, err := compressor.CompressFilesInDir(srcDir)
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(tgzName)

			_, _, _, err = cmdRunner.RunCommand("tar", "-xzpf", tgzName, "-C", dstDir)
			Expect(err).ToNot(HaveOccurred())

			content, err := fs.ReadFileString(dstDir + "/app.stdout.log")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(ContainSubstring("this is app stdout"))

			content, err = fs.ReadFileString(dstDir + "/app.stderr.log")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(ContainSubstring("this is app stderr"))

			content, err = fs.ReadFileString(dstDir + "/other_logs/other_app.stdout.log")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(ContainSubstring("this is other app stdout"))
		})
	})

	Describe("DecompressFileToDir", func() {
		It("decompresses the file to the given directory", func() {
			err := compressor.DecompressFileToDir(fixtureSrcTgz(), dstDir, CompressorOptions{})
			Expect(err).ToNot(HaveOccurred())

			content, err := fs.ReadFileString(dstDir + "/not-nested-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(ContainSubstring("not-nested-file"))

			content, err = fs.ReadFileString(dstDir + "/dir/nested-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(ContainSubstring("nested-file"))

			content, err = fs.ReadFileString(dstDir + "/dir/nested-dir/double-nested-file")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(ContainSubstring("double-nested-file"))

			Expect(dstDir + "/empty-dir").To(BeDir())
			Expect(dstDir + "/dir/empty-nested-dir").To(BeDir())
		})

		It("returns error if the destination does not exist", func() {
			fs.RemoveAll(dstDir)

			err := compressor.DecompressFileToDir(fixtureSrcTgz(), dstDir, CompressorOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(dstDir))
		})

		It("uses no same owner option", func() {
			cmdRunner := fakesys.NewFakeCmdRunner()
			compressor := NewTarballCompressor(cmdRunner, fs)

			tarballPath := fixtureSrcTgz()
			err := compressor.DecompressFileToDir(tarballPath, dstDir, CompressorOptions{})
			Expect(err).ToNot(HaveOccurred())

			Expect(1).To(Equal(len(cmdRunner.RunCommands)))
			Expect(cmdRunner.RunCommands[0]).To(Equal(
				[]string{
					"tar", "--no-same-owner",
					"-xzvf", tarballPath,
					"-C", dstDir,
				},
			))
		})

		It("uses same owner option", func() {
			cmdRunner := fakesys.NewFakeCmdRunner()
			compressor := NewTarballCompressor(cmdRunner, fs)

			tarballPath := fixtureSrcTgz()
			err := compressor.DecompressFileToDir(
				tarballPath,
				dstDir,
				CompressorOptions{SameOwner: true},
			)
			Expect(err).ToNot(HaveOccurred())

			Expect(1).To(Equal(len(cmdRunner.RunCommands)))
			Expect(cmdRunner.RunCommands[0]).To(Equal(
				[]string{
					"tar", "--same-owner",
					"-xzvf", tarballPath,
					"-C", dstDir,
				},
			))
		})
	})

	Describe("CleanUp", func() {
		It("removes tarball path", func() {
			fs := fakesys.NewFakeFileSystem()
			compressor := NewTarballCompressor(cmdRunner, fs)

			err := fs.WriteFileString("/fake-tarball.tar", "")
			Expect(err).ToNot(HaveOccurred())

			err = compressor.CleanUp("/fake-tarball.tar")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/fake-tarball.tar")).To(BeFalse())
		})

		It("returns error if removing tarball path fails", func() {
			fs := fakesys.NewFakeFileSystem()
			compressor := NewTarballCompressor(cmdRunner, fs)

			fs.RemoveAllError = errors.New("fake-remove-all-err")

			err := compressor.CleanUp("/fake-tarball.tar")
			Expect(err).To(MatchError("fake-remove-all-err"))
		})
	})
})
