package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
)

var _ = Describe("FileBytesArg", func() {
	Describe("UnmarshalFlag", func() {
		var (
			fs  *fakesys.FakeFileSystem
			arg FileBytesArg
		)

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			arg = FileBytesArg{FS: fs}
		})

		It("sets path and bytes", func() {
			fs.WriteFileString("/some/path", "content")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Path).To(Equal("/some/path"))
			Expect(arg.Bytes).To(Equal([]byte("content")))
		})

		It("returns an error if expanding path fails", func() {
			fs.ExpandPathErr = errors.New("fake-err")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns an error if reading file fails", func() {
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
