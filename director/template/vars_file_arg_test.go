package template_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/director/template"
)

var _ = Describe("VarsFileArg", func() {
	Describe("UnmarshalFlag", func() {
		var (
			fs  *fakesys.FakeFileSystem
			arg VarsFileArg
		)

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			arg = VarsFileArg{FS: fs}
		})

		It("sets read vars", func() {
			fs.WriteFileString("/some/path", "name1: var1\nname2: var2")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Vars).To(Equal(Variables{
				"name1": "var1",
				"name2": "var2",
			}))
		})

		It("returns an error if reading file fails", func() {
			fs.WriteFileString("/some/path", "content")
			fs.ReadFileError = errors.New("fake-err")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns an error if parsing file fails", func() {
			fs.WriteFileString("/some/path", "content")
			fs.ReadFileError = errors.New("fake-err")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns an error when it's empty", func() {
			err := (&arg).UnmarshalFlag("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected file path to be non-empty"))
		})
	})
})
