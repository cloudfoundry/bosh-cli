package cmd_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
)

var _ = Describe("TempRootConfigurator", func() {
	Describe("PrepareAndSetUpTempRoot", func() {
		var fs boshsys.FileSystem
		var testTempDir string
		var tempRoot string
		var logger boshlog.Logger

		BeforeEach(func() {
			var err error
			testTempDir, err = os.MkdirTemp("", "temp_root_configurator_test")
			Expect(err).ToNot(HaveOccurred())

			tempRoot = filepath.Join(testTempDir, "my-temp-root")
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fs = boshsys.NewOsFileSystem(logger)
		})

		AfterEach(func() {
			err := os.RemoveAll(testTempDir)
			Expect(err).ToNot(HaveOccurred())
		})

		var expectTempFileToBeCreatedUnderRoot = func(root, prefix string, fs boshsys.FileSystem) {
			file, err := fs.TempFile(prefix)
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(file.Name()) //nolint:errcheck

			Expect(file.Name()).To(HavePrefix(filepath.Join(root, prefix)))
		}

		Context("when the temp root already exists", func() {
			var existingFilePath string

			BeforeEach(func() {
				err := os.MkdirAll(tempRoot, os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
				existingFilePath = filepath.Join(tempRoot, "existing-file")
				err = os.WriteFile(existingFilePath, []byte{}, os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
			})

			It("clears out any files already in the temp directory", func() {
				tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

				Expect(existingFilePath).To(BeAnExistingFile())

				err := tempRootConfigurator.PrepareAndSetTempRoot(tempRoot, logger)
				Expect(err).ToNot(HaveOccurred())

				Expect(existingFilePath).ToNot(BeAnExistingFile())
			})

			It("sets the filesystem temp root", func() {
				tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

				err := tempRootConfigurator.PrepareAndSetTempRoot(tempRoot, logger)
				Expect(err).ToNot(HaveOccurred())

				expectTempFileToBeCreatedUnderRoot(tempRoot, "my-temp-file", fs)
			})

			It("returns an error if changing the temp root fails", func() {
				tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

				err := tempRootConfigurator.PrepareAndSetTempRoot("/dev/null/foo", logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("/dev/null"))
			})
		})

		Context("when the temp root doesn't exist", func() {
			It("sets the FileSystem temp root", func() {
				tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

				err := tempRootConfigurator.PrepareAndSetTempRoot(tempRoot, logger)
				Expect(err).ToNot(HaveOccurred())

				expectTempFileToBeCreatedUnderRoot(tempRoot, "my-temp-file", fs)
			})

			It("returns an error if changing the temp root fails", func() {
				tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

				err := tempRootConfigurator.PrepareAndSetTempRoot("/dev/null/foo", logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("/dev/null"))
			})
		})
	})
})
