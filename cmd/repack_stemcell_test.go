package cmd_test

import (
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	. "github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cloudfoundry/bosh-cli/stemcell/stemcellfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("RepackStemcellCmd", func() {
	var (
		fs        *fakesys.FakeFileSystem
		ui        *fakeui.FakeUI
		command   RepackStemcellCmd
		extractor *stemcellfakes.FakeExtractor
		packer    *stemcellfakes.FakePacker
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}

		fs.WriteFileString("intermediate-stemcell.tgz", "some-content")

		extractor = stemcellfakes.NewFakeExtractor()
		packer = new(stemcellfakes.FakePacker)
		command = NewRepackStemcellCmd(ui, fs, extractor, packer)
	})

	Describe("Run", func() {
		var (
			opts RepackStemcellOpts
		)

		BeforeEach(func() {
			opts = RepackStemcellOpts{}
		})

		act := func() error { return command.Run(opts) }

		FContext("when url is a local file (file or no prefix)", func() {
			var (
				extractedStemcell *stemcellfakes.FakeExtractedStemcell
				err               error
			)

			BeforeEach(func() {
				opts.Args.PathToStemcell = "some-stemcell.tgz"
				opts.Args.PathToResult = "repacked-stemcell.tgz"
				extractedStemcell = &stemcellfakes.FakeExtractedStemcell{}
			})

			It("duplicates the stemcell and saves to PathToResult", func() {
				extractor.SetExtractBehavior("some-stemcell.tgz", extractedStemcell, nil)

				packer.PackReturns("intermediate-stemcell.tgz", nil)
				err = act()
				Expect(err).ToNot(HaveOccurred())

				Expect(len(extractor.ExtractInputs)).To(Equal(1))
				Expect(extractor.ExtractInputs[0].TarballPath).To(Equal("some-stemcell.tgz"))

				Expect(packer.PackCallCount()).To(Equal(1))
				Expect(packer.PackArgsForCall(0)).To(Equal(extractedStemcell))

				Expect(fs.FileExists("intermediate-stemcell.tgz")).To(BeFalse())
				Expect(fs.FileExists("repacked-stemcell.tgz")).To(BeTrue())
			})

			Context("when error ocurrs", func() {
				Context("when the destination file already exists", func() {
					BeforeEach(func() {
						fs.WriteFileString("repacked-stemcell.tgz", "I already exist -- ha!")
						extractor.SetExtractBehavior("some-stemcell.tgz", extractedStemcell, nil)
						err = act()
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("destination file exists"))
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
						packer.PackReturns("", errors.New("fake-error"))
						err = act()
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())

						Expect(len(extractor.ExtractInputs)).To(Equal(1))
						Expect(extractor.ExtractInputs[0].TarballPath).To(Equal("some-stemcell.tgz"))

						Expect(packer.PackCallCount()).To(Equal(1))
					})
				})

				Context("when it fails to move the new stemcell into given file path", func() {
					BeforeEach(func() {
						extractor.SetExtractBehavior("some-stemcell.tgz", extractedStemcell, nil)
						packer.PackReturns("intermediate-stemcell.tgz", nil)
						fs.RenameError = errors.New("fake-error")
						err = act()
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())

						Expect(len(extractor.ExtractInputs)).To(Equal(1))
						Expect(extractor.ExtractInputs[0].TarballPath).To(Equal("some-stemcell.tgz"))

						Expect(packer.PackCallCount()).To(Equal(1))
						Expect(packer.PackArgsForCall(0)).To(Equal(extractedStemcell))
					})
				})
			})
		})
	})
})
