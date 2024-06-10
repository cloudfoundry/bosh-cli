package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("AddBlobCmd", func() {
	var (
		blobsDir *fakereldir.FakeBlobsDir
		fs       *fakesys.FakeFileSystem
		ui       *fakeui.FakeUI
		command  cmd.AddBlobCmd
	)

	BeforeEach(func() {
		blobsDir = &fakereldir.FakeBlobsDir{}
		fs = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}
		command = cmd.NewAddBlobCmd(blobsDir, fs, ui)
	})

	Describe("Run", func() {
		var (
			addBlobOpts opts.AddBlobOpts
		)

		BeforeEach(func() {
			err := fs.WriteFileString("/path/to/blob.tgz", "blob")
			Expect(err).ToNot(HaveOccurred())
			addBlobOpts = opts.AddBlobOpts{
				Args: opts.AddBlobArgs{
					Path:      "/path/to/blob.tgz",
					BlobsPath: "my-blob.tgz",
				},
			}
		})

		act := func() error { return command.Run(addBlobOpts) }

		It("starts tracking blob", func() {
			blobsDir.TrackBlobReturns(boshreldir.Blob{Path: "my-blob.tgz"}, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.TrackBlobCallCount()).To(Equal(1))

			blobsPath, src := blobsDir.TrackBlobArgsForCall(0)
			Expect(blobsPath).To(Equal("my-blob.tgz"))

			file := src.(*fakesys.FakeFile)
			Expect(file.Name()).To(Equal("/path/to/blob.tgz"))
			Expect(file.Stats.Open).To(BeFalse())

			Expect(ui.Said).To(Equal([]string{"Added blob 'my-blob.tgz'"}))
		})

		It("returns error if tracking fails", func() {
			blobsDir.TrackBlobReturns(boshreldir.Blob{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(ui.Said).To(BeEmpty())
		})

		It("returns error if file cannot be open", func() {
			fs.OpenFileErr = errors.New("fake-err")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(blobsDir.TrackBlobCallCount()).To(Equal(0))

			Expect(ui.Said).To(BeEmpty())
		})
	})
})
