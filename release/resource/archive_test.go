package resource_test

import (
	"errors"
	"fmt"
	"os"

	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	fakecmd "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	fakecrypto "github.com/cloudfoundry/bosh-cli/crypto/fakes"
	. "github.com/cloudfoundry/bosh-cli/release/resource"
	fakeres "github.com/cloudfoundry/bosh-cli/release/resource/resourcefakes"
)

var _ = Describe("Archive", func() {
	var (
		archive Archive
	)

	BeforeEach(func() {
		archive = nil
	})

	Describe("Fingerprint", func() {
		var (
			fingerprinter *fakeres.FakeFingerprinter
			sha1calc      *fakecrypto.FakeSha1Calculator
			compressor    *fakecmd.FakeCompressor
			cmdRunner     *fakesys.FakeCmdRunner
			fs            *fakesys.FakeFileSystem
		)

		BeforeEach(func() {
			releaseDirPath := "/tmp/release"
			fingerprinter = &fakeres.FakeFingerprinter{}
			sha1calc = fakecrypto.NewFakeSha1Calculator()
			compressor = fakecmd.NewFakeCompressor()
			cmdRunner = fakesys.NewFakeCmdRunner()
			fs = fakesys.NewFakeFileSystem()
			archive = NewArchiveImpl(
				[]File{NewFile("/tmp/file", "/tmp")},
				[]File{NewFile("/tmp/prep-file", "/tmp")},
				[]string{"chunk"},
				releaseDirPath,
				fingerprinter,
				compressor,
				sha1calc,
				cmdRunner,
				fs,
			)
		})

		It("returns fingerprint", func() {
			fingerprinter.CalculateReturns("fp", nil)

			fp, err := archive.Fingerprint()
			Expect(err).ToNot(HaveOccurred())
			Expect(fp).To(Equal("fp"))

			files, chunks := fingerprinter.CalculateArgsForCall(0)
			Expect(files).To(Equal([]File{NewFile("/tmp/file", "/tmp")}))
			Expect(chunks).To(Equal([]string{"chunk"}))
		})

		It("returns error", func() {
			fingerprinter.CalculateReturns("", errors.New("fake-err"))

			_, err := archive.Fingerprint()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Build", func() {
		var (
			uniqueDir string
			fs        boshsys.FileSystem

			compressor boshcmd.Compressor
			sha1calc   bicrypto.SHA1Calculator
		)

		BeforeEach(func() {
			releaseDirPath := "/tmp/release"

			suffix, err := boshuuid.NewGenerator().Generate()
			Expect(err).ToNot(HaveOccurred())

			uniqueDir = "/tmp/" + suffix

			logger := boshlog.NewLogger(boshlog.LevelNone)
			fs = boshsys.NewOsFileSystemWithStrictTempRoot(logger)

			err = fs.ChangeTempRoot(uniqueDir)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(uniqueDir+"/file1", "file1")
			Expect(err).ToNot(HaveOccurred())

			err = fs.MkdirAll(uniqueDir+"/dir", os.FileMode(0744))
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(uniqueDir+"/dir/file2", "file2")
			Expect(err).ToNot(HaveOccurred())

			err = fs.Chmod(uniqueDir+"/dir/file2", os.FileMode(0745))
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(uniqueDir+"/dir/file3", "file3")
			Expect(err).ToNot(HaveOccurred())

			err = fs.MkdirAll(uniqueDir+"/dir/symlink-dir-target", os.FileMode(0744))
			Expect(err).ToNot(HaveOccurred())

			err = fs.Symlink("symlink-dir-target", uniqueDir+"/dir/symlink-dir")
			Expect(err).ToNot(HaveOccurred())

			err = fs.Symlink("../file1", uniqueDir+"/dir/symlink-file")
			Expect(err).ToNot(HaveOccurred())

			err = fs.Symlink("nonexistant-file", uniqueDir+"/dir/symlink-file-missing")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(uniqueDir+"/run-build-dir", "echo -n $BUILD_DIR > build-dir")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(uniqueDir+"/run-release-dir", "echo -n $RELEASE_DIR > release-dir")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(uniqueDir+"/run-file3", "rm dir/file3")
			Expect(err).ToNot(HaveOccurred())

			sha1calc = bicrypto.NewSha1Calculator(fs)
			fingerprinter := NewFingerprinterImpl(sha1calc, fs)
			cmdRunner := boshsys.NewExecCmdRunner(logger)
			compressor = boshcmd.NewTarballCompressor(cmdRunner, fs)

			archive = NewArchiveImpl(
				[]File{
					NewFile(uniqueDir+"/file1", uniqueDir),
					NewFile(uniqueDir+"/dir/file2", uniqueDir),
					NewFile(uniqueDir+"/dir/file3", uniqueDir),
					NewFile(uniqueDir+"/dir/symlink-file", uniqueDir),
					NewFile(uniqueDir+"/dir/symlink-file-missing", uniqueDir),
					NewFile(uniqueDir+"/dir/symlink-dir", uniqueDir),
				},
				[]File{
					NewFile(uniqueDir+"/run-build-dir", uniqueDir),
					NewFile(uniqueDir+"/run-release-dir", uniqueDir),
					NewFile(uniqueDir+"/run-file3", uniqueDir),
				},
				[]string{"chunk"},
				releaseDirPath,
				fingerprinter,
				compressor,
				sha1calc,
				cmdRunner,
				fs,
			)
		})

		AfterEach(func() {
			if fs != nil {
				_ = fs.RemoveAll(uniqueDir)
			}
		})

		modeAsStr := func(m os.FileMode) string {
			return fmt.Sprintf("%#o", m)
		}

		It("returns archive, sha1 when built successfully", func() {
			archivePath, archiveSHA1, err := archive.Build("31a86e1b2b76e47ca5455645bb35018fe7f73e5d")
			Expect(err).ToNot(HaveOccurred())

			actualArchiveSHA1, err := sha1calc.Calculate(archivePath)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualArchiveSHA1).To(Equal(archiveSHA1))

			decompPath, err := fs.TempDir("test-resource")
			Expect(err).ToNot(HaveOccurred())

			err = compressor.DecompressFileToDir(archivePath, decompPath, boshcmd.CompressorOptions{})
			Expect(err).ToNot(HaveOccurred())

			{
				// Copies specified files
				Expect(fs.ReadFileString(decompPath + "/file1")).To(Equal("file1"))
				Expect(fs.ReadFileString(decompPath + "/dir/file2")).To(Equal("file2"))

				// Copies specified symlinks
				stat, err := fs.Lstat(decompPath + "/dir/symlink-file")
				Expect(err).ToNot(HaveOccurred())
				Expect(stat.Mode()&os.ModeSymlink != 0).To(BeTrue())
				Expect(fs.Readlink(decompPath + "/dir/symlink-file")).To(Equal("../file1"))

				stat, err = fs.Lstat(decompPath + "/dir/symlink-file-missing")
				Expect(err).ToNot(HaveOccurred())
				Expect(stat.Mode()&os.ModeSymlink != 0).To(BeTrue())
				Expect(fs.Readlink(decompPath + "/dir/symlink-file-missing")).To(Equal("nonexistant-file"))

				stat, err = fs.Lstat(decompPath + "/dir/symlink-dir")
				Expect(err).ToNot(HaveOccurred())
				Expect(stat.Mode()&os.ModeSymlink != 0).To(BeTrue())
				Expect(fs.Readlink(decompPath + "/dir/symlink-dir")).To(Equal("symlink-dir-target"))
				Expect(fs.FileExists(decompPath + "/dir/simlink-dir-target")).To(BeFalse())

				// Dir permissions
				stat, err = fs.Stat(decompPath + "/dir")
				Expect(err).ToNot(HaveOccurred())
				Expect(modeAsStr(stat.Mode())).To(Equal("020000000744")) // 02... is for directory

				// File permissions
				stat, err = fs.Stat(decompPath + "/dir/file2")
				Expect(err).ToNot(HaveOccurred())
				Expect(modeAsStr(stat.Mode())).To(Equal("0745"))
			}

			{
				// Runs scripts
				Expect(fs.ReadFileString(decompPath + "/build-dir")).ToNot(BeEmpty())
				Expect(fs.ReadFileString(decompPath + "/release-dir")).To(Equal("/tmp/release"))
				Expect(fs.FileExists(decompPath + "/dir/file3")).To(BeFalse())
			}

			{
				// Deletes scripts
				Expect(fs.FileExists(decompPath + "/run-build-dir")).To(BeFalse())
				Expect(fs.FileExists(decompPath + "/run-release-dir")).To(BeFalse())
			}
		})
	})
})
