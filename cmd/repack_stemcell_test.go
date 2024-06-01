package cmd_test

import (
	"errors"

	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	"github.com/cloudfoundry/bosh-cli/v7/stemcell/stemcellfakes"
)

var _ = Describe("RepackStemcellCmd", func() {
	Describe("Run", func() {
		var (
			fakeExtractor *stemcellfakes.FakeExtractor
			command       cmd.RepackStemcellCmd

			repackStemcellOpts opts.RepackStemcellOpts

			stemcellPath string
			resultPath   string

			fakeExtractedStemcell *stemcellfakes.FakeExtractedStemcell
		)

		BeforeEach(func() {
			fakeExtractor = stemcellfakes.NewFakeExtractor()
			command = cmd.NewRepackStemcellCmd(fakeExtractor)

			stemcellPath = "definitely/a/path"
			resultPath = "a/different/path/here"
			repackStemcellOpts = opts.RepackStemcellOpts{
				Args: opts.RepackStemcellArgs{
					PathToStemcell: stemcellPath,
					PathToResult:   opts.FileArg{ExpandedPath: resultPath},
				},
			}

			fakeExtractedStemcell = &stemcellfakes.FakeExtractedStemcell{}
			fakeExtractor.SetExtractBehavior(stemcellPath, fakeExtractedStemcell, nil)
		})

		Context("no flags are passed", func() {
			It("does not modify the extracted stemcell before packing", func() {
				err := command.Run(repackStemcellOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.SetNameCallCount()).To(BeZero())
				Expect(fakeExtractedStemcell.SetVersionCallCount()).To(BeZero())
				Expect(fakeExtractedStemcell.EmptyImageCallCount()).To(BeZero())
				Expect(fakeExtractedStemcell.SetCloudPropertiesCallCount()).To(BeZero())
				Expect(fakeExtractedStemcell.SetFormatCallCount()).To(BeZero())

				Expect(fakeExtractedStemcell.PackCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.PackArgsForCall(0)).To(Equal(resultPath))
			})
		})

		Context("name flag is passed", func() {
			var name string

			BeforeEach(func() {
				name = "foo"
				repackStemcellOpts.Name = name
			})

			It("modifies the extracted stemcell's name", func() {
				err := command.Run(repackStemcellOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.SetNameCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.SetNameArgsForCall(0)).To(Equal(name))
				Expect(fakeExtractedStemcell.PackArgsForCall(0)).To(Equal(resultPath))
			})
		})

		Context("version flag is passed", func() {
			var version string

			BeforeEach(func() {
				version = "1.2.3"
				repackStemcellOpts.Version = version
			})

			It("modifies the extracted stemcell's version", func() {
				err := command.Run(repackStemcellOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.SetVersionCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.SetVersionArgsForCall(0)).To(Equal(version))
				Expect(fakeExtractedStemcell.PackArgsForCall(0)).To(Equal(resultPath))
			})
		})

		Context("empty-image flag is passed", func() {
			BeforeEach(func() {
				repackStemcellOpts.EmptyImage = true
			})

			It("modifies the extracted stemcell's image", func() {
				err := command.Run(repackStemcellOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.EmptyImageCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.PackArgsForCall(0)).To(Equal(resultPath))
			})

			It("returns an error if empty image fails", func() {
				fakeExtractedStemcell.EmptyImageReturns(errors.New("uh oh"))

				err := command.Run(repackStemcellOpts)
				Expect(err).To(MatchError("uh oh"))

				Expect(fakeExtractedStemcell.EmptyImageCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.PackCallCount()).To(BeZero())
			})
		})

		Context("cloud properties flag is passed", func() {
			var cloudProperties string

			BeforeEach(func() {
				cloudProperties = `{"foo": "bar"}`
				repackStemcellOpts.CloudProperties = cloudProperties
			})

			It("modifies the extracted stemcell's cloud properties", func() {
				err := command.Run(repackStemcellOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.SetCloudPropertiesCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.SetCloudPropertiesArgsForCall(0)).To(Equal(biproperty.Map{"foo": "bar"}))
				Expect(fakeExtractedStemcell.PackArgsForCall(0)).To(Equal(resultPath))
			})

			It("returns an error if yaml umarshalling fails", func() {
				repackStemcellOpts.CloudProperties = "not/yaml/no/siree"

				err := command.Run(repackStemcellOpts)
				Expect(err).To(MatchError(ContainSubstring("yaml: unmarshal")))

				Expect(fakeExtractedStemcell.SetCloudPropertiesCallCount()).To(BeZero())
				Expect(fakeExtractedStemcell.PackCallCount()).To(BeZero())
			})
		})

		Context("format flag is passed", func() {
			var format []string

			BeforeEach(func() {
				format = []string{"foo", "bar", "baz"}
				repackStemcellOpts.Format = format
			})

			It("modifies the extracted stemcell's format", func() {
				err := command.Run(repackStemcellOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeExtractedStemcell.SetFormatCallCount()).To(Equal(1))
				Expect(fakeExtractedStemcell.SetFormatArgsForCall(0)).To(Equal(format))
				Expect(fakeExtractedStemcell.PackArgsForCall(0)).To(Equal(resultPath))
			})
		})
	})
})
