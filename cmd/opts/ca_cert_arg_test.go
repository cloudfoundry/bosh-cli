package opts_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

var _ = Describe("CACertArg", func() {
	Describe("UnmarshalFlag", func() {
		var (
			fs  *fakesys.FakeFileSystem
			arg CACertArg
		)

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			arg = CACertArg{FS: fs}
		})

		It("sets bytes from value if value is a PEM certificate (contains BEGIN)", func() {
			err := (&arg).UnmarshalFlag("BEGIN ...")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Content).To(Equal("BEGIN ..."))
		})

		It("sets bytes from file contents if value is not a PEM certificate", func() {
			err := fs.WriteFileString("/some/path", "content")
			Expect(err).ToNot(HaveOccurred())

			err = (&arg).UnmarshalFlag("/some/path")
			Expect(err).ToNot(HaveOccurred())
			Expect(arg.Content).To(Equal("content"))
		})

		It("returns an error if expanding path fails", func() {
			fs.ExpandPathErr = errors.New("fake-err")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns an error if reading file fails", func() {
			err := fs.WriteFileString("/some/path", "content")
			Expect(err).ToNot(HaveOccurred())
			fs.ReadFileError = errors.New("fake-err")

			err = (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns an error when it's empty", func() {
			err := (&arg).UnmarshalFlag("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected CA cert to be non-empty"))
		})
	})
})
