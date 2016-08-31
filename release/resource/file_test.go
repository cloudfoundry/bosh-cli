package resource_test

import (
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/release/resource"
)

var _ = Describe("NewFile", func() {
	It("returns file with relative path that does not start with separator", func() {
		file := NewFile("/tmp/file", "/tmp")
		Expect(file.Path).To(Equal("/tmp/file"))
		Expect(file.DirPath).To(Equal("/tmp"))
		Expect(file.RelativePath).To(Equal("file"))
	})

	It("returns file with relative path when dir path ends with separator", func() {
		file := NewFile("/tmp/file", "/tmp/")
		Expect(file.Path).To(Equal("/tmp/file"))
		Expect(file.DirPath).To(Equal("/tmp"))
		Expect(file.RelativePath).To(Equal("file"))
	})
})

var _ = Describe("File", func() {
	Describe("WithNewDir", func() {
		It("returns file as if it was from a different dir", func() {
			file := NewFile("/tmp/file", "/tmp/").WithNewDir("/other")
			Expect(file.Path).To(Equal("/other/file"))
			Expect(file.DirPath).To(Equal("/other"))
			Expect(file.RelativePath).To(Equal("file"))
		})
	})
})

var _ = Describe("FileRelativePathSorting", func() {
	It("sorts files based on relative path", func() {
		file2 := NewFile("/tmp/file2", "/tmp/")
		file1 := NewFile("/tmp/file1", "/tmp/")
		file := NewFile("/tmp/file", "/tmp/")
		files := []File{file2, file1, file}
		sort.Sort(FileRelativePathSorting(files))
		Expect(files).To(Equal([]File{file, file1, file2}))
	})
})
