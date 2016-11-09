package cmd_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	fakecrypto "github.com/cloudfoundry/bosh-cli/crypto/fakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/clock/fakeclock"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
)

var _ = Describe("UIDownloader", func() {
	var (
		director    *fakedir.FakeDirector
		sha1calc    *fakecrypto.FakeSha1Calculator
		fs          *fakesys.FakeFileSystem
		timeService clock.Clock
		ui          *fakeui.FakeUI
		downloader  UIDownloader
	)

	BeforeEach(func() {
		director = &fakedir.FakeDirector{}
		sha1calc = fakecrypto.NewFakeSha1Calculator()
		timeService = fakeclock.NewFakeClock(time.Date(2009, time.November, 10, 23, 1, 2, 333, time.UTC))
		fs = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}
		downloader = NewUIDownloader(director, sha1calc, timeService, fs, ui)
	})

	Describe("Download", func() {
		var expectedPath string

		BeforeEach(func() {
			expectedPath = "/fake-dst-dir/prefix-20091110-230102-000000333.tgz"

			err := fs.MkdirAll("/fake-dst-dir", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
		})

		itReturnsErrs := func(act func() error) {
			It("returns error if downloading resource fails", func() {
				err := fs.MkdirAll("/fake-dst-dir", os.ModePerm)
				Expect(err).ToNot(HaveOccurred())

				fs.ReturnTempFile = fakesys.NewFakeFile("/some-tmp-file", fs)

				director.DownloadResourceUncheckedStub = func(_ string, _ io.Writer) error {
					return errors.New("fake-err")
				}

				err = act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(fs.FileExists("/some-tmp-file")).To(BeFalse())
				Expect(fs.FileExists(expectedPath)).To(BeFalse())
			})

			It("returns error if temp file cannot be created", func() {
				fs.TempFileError = errors.New("fake-err")

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.DownloadResourceUncheckedCallCount()).To(Equal(0))
				Expect(fs.FileExists(expectedPath)).To(BeFalse())
			})
		}

		Context("when SHA1 is provided", func() {
			act := func() error { return downloader.Download("fake-blob-id", "fake-sha1", "prefix", "/fake-dst-dir") }

			It("downloads specified blob to a specific destination", func() {
				fs.ReturnTempFile = fakesys.NewFakeFile("/some-tmp-file", fs)

				sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					"/some-tmp-file": fakecrypto.CalculateInput{Sha1: "fake-sha1"},
				})

				director.DownloadResourceUncheckedStub = func(_ string, out io.Writer) error {
					out.Write([]byte("content"))
					return nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("/some-tmp-file")).To(BeFalse())
				Expect(fs.FileExists(expectedPath)).To(BeTrue())
				Expect(fs.ReadFileString(expectedPath)).To(Equal("content"))

				blobID, _ := director.DownloadResourceUncheckedArgsForCall(0)
				Expect(blobID).To(Equal("fake-blob-id"))

				Expect(ui.Said).To(Equal([]string{
					fmt.Sprintf("Downloading resource 'fake-blob-id' to '%s'...", expectedPath)}))
			})

			It("returns error if sha1 does not match expected sha1", func() {
				fs.ReturnTempFile = fakesys.NewFakeFile("/some-tmp-file", fs)

				sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					"/some-tmp-file": fakecrypto.CalculateInput{Sha1: "non-matching-sha1"},
				})

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected file SHA1 to be 'fake-sha1' but was 'non-matching-sha1'"))

				Expect(fs.FileExists(expectedPath)).To(BeFalse())
			})

			It("returns error if sha1 check fails", func() {
				fs.ReturnTempFile = fakesys.NewFakeFile("/some-tmp-file", fs)

				sha1calc.SetCalculateBehavior(map[string]fakecrypto.CalculateInput{
					"/some-tmp-file": fakecrypto.CalculateInput{Err: errors.New("fake-err")},
				})

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(fs.FileExists(expectedPath)).To(BeFalse())
			})

			itReturnsErrs(act)
		})

		Context("when SHA1 is not provided", func() {
			act := func() error { return downloader.Download("fake-blob-id", "", "prefix", "/fake-dst-dir") }

			It("downloads specified blob to a specific destination without checking SHA1", func() {
				fs.ReturnTempFile = fakesys.NewFakeFile("/some-tmp-file", fs)

				director.DownloadResourceUncheckedStub = func(_ string, out io.Writer) error {
					out.Write([]byte("content"))
					return nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("/some-tmp-file")).To(BeFalse())
				Expect(fs.FileExists(expectedPath)).To(BeTrue())
				Expect(fs.ReadFileString(expectedPath)).To(Equal("content"))

				blobID, _ := director.DownloadResourceUncheckedArgsForCall(0)
				Expect(blobID).To(Equal("fake-blob-id"))

				Expect(ui.Said).To(Equal([]string{
					fmt.Sprintf("Downloading resource 'fake-blob-id' to '%s'...", expectedPath)}))
			})

			itReturnsErrs(act)
		})

		Context("when downloading across devices", func() {
			BeforeEach(func() {
				fs.RenameError = &os.LinkError{
					Err: syscall.Errno(0x12),
				}
			})

			act := func() error { return downloader.Download("fake-blob-id", "", "prefix", "/fake-dst-dir") }

			It("downloads specified blob to a specific destination without checking SHA1", func() {
				fs.ReturnTempFile = fakesys.NewFakeFile("/some-tmp-file", fs)

				director.DownloadResourceUncheckedStub = func(_ string, out io.Writer) error {
					out.Write([]byte("content"))
					return nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("/some-tmp-file")).To(BeFalse())
				Expect(fs.FileExists(expectedPath)).To(BeTrue())
				Expect(fs.ReadFileString(expectedPath)).To(Equal("content"))

				blobID, _ := director.DownloadResourceUncheckedArgsForCall(0)
				Expect(blobID).To(Equal("fake-blob-id"))

				Expect(ui.Said).To(Equal([]string{
					fmt.Sprintf("Downloading resource 'fake-blob-id' to '%s'...", expectedPath)}))
			})

			itReturnsErrs(act)
		})
	})
})
