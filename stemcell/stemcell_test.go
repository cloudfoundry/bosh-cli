package stemcell_test

import (
	. "github.com/cloudfoundry/bosh-cli/stemcell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("Stemcell", func() {
	var (
		stemcell      ExtractedStemcell
		manifest      Manifest
		fakefs        *fakesys.FakeFileSystem
		extractedPath string
	)

	BeforeEach(func() {
		manifest = Manifest{}

		extractedPath = "fake-path"

		fakefs = fakesys.NewFakeFileSystem()

		stemcell = NewExtractedStemcell(
			manifest,
			extractedPath,
			fakefs,
		)
	})

	Describe("#Manifest", func() {
		It("returns the manifest", func() {
			Expect(stemcell.Manifest()).To(Equal(manifest))
		})
	})

	Describe("Delete", func() {
		var removeAllCalled bool

		BeforeEach(func() {
			fakefs.RemoveAllStub = func(_ string) error {
				removeAllCalled = true
				return nil
			}
		})

		It("removes the extrated path", func() {
			Expect(stemcell.Delete()).To(BeNil())
			Expect(removeAllCalled).To(BeTrue())
		})
	})

	Describe("String", func() {
		BeforeEach(func() {
			manifest = Manifest{
				Name:    "some-name",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)
		})

		It("returns the name and the version", func() {
			Expect(stemcell.String()).To(Equal("ExtractedStemcell{name=some-name version=some-version}"))
		})
	})

	Describe("OsAndVersion", func() {
		BeforeEach(func() {
			manifest = Manifest{
				OS:      "some-os",
				Version: "some-version",
			}

			stemcell = NewExtractedStemcell(
				manifest,
				extractedPath,
				fakefs,
			)
		})

		It("returns the name and the version", func() {
			Expect(stemcell.OsAndVersion()).To(Equal("some-os/some-version"))
		})
	})
})
