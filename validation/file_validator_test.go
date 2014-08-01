package validation_test

import (
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/validation"
)

var _ = Describe("fileValidator", func() {
	var (
		fakeFs        *fakesys.FakeFileSystem
		fileValidator FileValidator
	)
	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		fileValidator = NewFileValidator(fakeFs)
	})

	Describe("Exists", func() {
		Context("when the file at the given path exists", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString("/some/path", "")
			})

			It("does not return an error", func() {
				err := fileValidator.Exists("/some/path")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the file path does not exist", func() {
			It("returns err", func() {
				err := fileValidator.Exists("/some/path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Path '/some/path' does not exist"))
			})
		})
	})
})
