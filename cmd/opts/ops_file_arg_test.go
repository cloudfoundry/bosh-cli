package opts_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OpsFileArg", func() {
	Describe("UnmarshalFlag", func() {
		var (
			fs  *fakesys.FakeFileSystem
			arg OpsFileArg
		)

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			arg = OpsFileArg{FS: fs}
		})

		It("sets read operations", func() {
			fs.WriteFileString("/some/path", `
- type: remove
  path: /a
- type: remove
  path: /b
`)

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).ToNot(HaveOccurred())

			Expect(arg.Ops).To(Equal(patch.Ops{
				patch.DescriptiveOp{
					Op: patch.RemoveOp{
						Path: patch.MustNewPointerFromString("/a"),
					},
					ErrorMsg: "operation [0] in /some/path failed",
				},
				patch.DescriptiveOp{
					Op: patch.RemoveOp{
						Path: patch.MustNewPointerFromString("/b"),
					},
					ErrorMsg: "operation [1] in /some/path failed",
				},
			}))
		})

		It("returns an error if operations are not valid", func() {
			fs.WriteFileString("/some/path", "- type: unknown")

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`Building ops: Unknown operation [0] with type 'unknown' within
{
  "Type": "unknown",
  "Error": "operation [0] in /some/path failed"
}`))
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

			err := (&arg).UnmarshalFlag("/some/path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Deserializing ops file '/some/path'"))
		})

		It("returns an error when it's empty", func() {
			err := (&arg).UnmarshalFlag("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected file path to be non-empty"))
		})
	})
})
