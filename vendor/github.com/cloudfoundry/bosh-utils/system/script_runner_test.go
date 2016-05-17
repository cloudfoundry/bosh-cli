package system_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/cloudfoundry/bosh-utils/system"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("ScriptRunner", func() {
	var (
		cmdRunner    *fakesys.FakeCmdRunner
		fs           *fakesys.FakeFileSystem
		scriptRunner ScriptRunner
		logger       boshlog.Logger
	)

	BeforeEach(func() {
		scriptCommandFactory := &fakesys.FakeCommandFactory{}
		scriptCommandFactory.ReturnExtension = ".fake-ext"

		cmdRunner = fakesys.NewFakeCmdRunner()

		fs = fakesys.NewFakeFileSystem()

		var err error
		fs.ReturnTempFile, err = fs.OpenFile("/fake-temp-file", os.O_WRONLY, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		logger = boshlog.NewLogger(boshlog.LevelNone)
		scriptRunner = NewConcreteScriptRunner(scriptCommandFactory, cmdRunner, fs, logger)
	})

	Describe("RunCommand", func() {
		It("runs a successful script command and doesnt return an error", func() {
			script := `
Write-Output stdout
[Console]::Error.WriteLine('stderr')
`

			cmdRunner.AddCmdResult("/fake-temp-file.fake-ext", fakesys.FakeCmdResult{
				Stdout:     "stdout",
				Stderr:     "stderr",
				ExitStatus: 0,
			})

			var scriptContent string
			cmdCallback := func() {
				var err error
				scriptContent, err = fs.ReadFileString("/fake-temp-file.fake-ext")
				Expect(err).NotTo(HaveOccurred())
			}
			cmdRunner.SetCmdCallback("/fake-temp-file.fake-ext", cmdCallback)

			stdout, stderr, err := scriptRunner.Run(script)
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal("stdout"))
			Expect(stderr).To(Equal("stderr"))

			Expect(scriptContent).To(Equal(script))
		})

		It("runs a failing Powershell command and returns error", func() {
			cmdRunner.AddCmdResult("/fake-temp-file.fake-ext", fakesys.FakeCmdResult{
				Error: errors.New("failed"),
			})
			_, _, err := scriptRunner.Run("")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed"))
		})

		It("cleans temporary files", func() {
			_, _, err := scriptRunner.Run("")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/fake-temp-file")).To(BeFalse())
			Expect(fs.FileExists("/fake-temp-file.fake-ext")).To(BeFalse())
		})

		Context("filesystem errors", func() {
			Context("when creating Tempfile fails", func() {
				It("errors out", func() {
					fs.TempFileError = errors.New("boo")

					_, _, err := scriptRunner.Run("")
					Expect(err.Error()).To(Equal("Creating tempfile: boo"))
				})
			})

			Context("when writing to the Tempfile fails", func() {
				It("errors out", func() {
					fs.WriteFileError = errors.New("foo")

					_, _, err := scriptRunner.Run("")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Writing to tempfile: foo"))
				})
			})

			Context("when closing Tempfile fails", func() {
				It("errors out", func() {
					tempFile := fs.ReturnTempFile.(*fakesys.FakeFile)
					tempFile.CloseErr = errors.New("fake-close-error")

					_, _, err := scriptRunner.Run("")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Closing tempfile: fake-close-error"))
				})
			})

			Context("when renaming Tempfile fails", func() {
				It("errors out", func() {
					fs.RenameError = errors.New("fake-rename-error")

					_, _, err := scriptRunner.Run("")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Renaming tempfile: fake-rename-error"))
				})
			})
		})
	})
})
