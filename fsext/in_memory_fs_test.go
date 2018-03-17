package fsext_test

import (
	"os"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/fsext"
)

var _ = Describe("InMemoryFS", func() {
	var (
		fs *InMemoryFS
	)

	BeforeEach(func() {
		fs = NewInMemoryFS()
	})

	Describe("ExpandPath", func() {
		It("returns original path", func() {
			path, err := fs.ExpandPath("test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("test-path.json"))
		})
	})

	Describe("ReadFile/ReadFileString/FileExists/WriteFile/WriteFileString", func() {
		It("saves and retrieves values", func() {
			_, err := fs.ReadFileString("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find 'test-key'"))

			found := fs.FileExists("test-key")
			Expect(found).To(BeFalse())

			err = fs.WriteFileString("test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			valBytes, err := fs.ReadFileWithOpts("test-key", boshsys.ReadOpts{})
			Expect(err).ToNot(HaveOccurred())
			Expect(valBytes).To(Equal([]byte("test-value")))

			found = fs.FileExists("test-key")
			Expect(found).To(BeTrue())

			err = fs.WriteFile("test-key", []byte("test-value2"))
			Expect(err).ToNot(HaveOccurred())

			val, err := fs.ReadFileString("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test-value2"))

			found = fs.FileExists("test-key")
			Expect(found).To(BeTrue())
		})
	})

	Describe("RemoveAll", func() {
		It("removes individual files", func() {
			err := fs.WriteFileString("test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			found := fs.FileExists("test-key")
			Expect(found).To(BeTrue())

			err = fs.RemoveAll("test-key")
			Expect(err).ToNot(HaveOccurred())

			found = fs.FileExists("test-key")
			Expect(found).To(BeFalse())
		})

		It("removes files that do not exist successfully", func() {
			err := fs.RemoveAll("test-key-does-not-exist")
			Expect(err).ToNot(HaveOccurred())
		})

		It("removes groups of files", func() {
			err := fs.WriteFileString("dir/test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("dir2/test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("dir-test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())

			found := fs.FileExists("dir/test-key")
			Expect(found).To(BeTrue())

			found = fs.FileExists("dir2/test-key")
			Expect(found).To(BeTrue())

			found = fs.FileExists("dir-test-key")
			Expect(found).To(BeTrue())

			err = fs.RemoveAll("dir")
			Expect(err).ToNot(HaveOccurred())

			found = fs.FileExists("dir/test-key")
			Expect(found).To(BeFalse())

			found = fs.FileExists("dir2/test-key")
			Expect(found).To(BeTrue())

			found = fs.FileExists("dir-test-key")
			Expect(found).To(BeTrue())

			err = fs.RemoveAll("dir2/")
			Expect(err).ToNot(HaveOccurred())

			found = fs.FileExists("dir2/test-key")
			Expect(found).To(BeFalse())
		})
	})

	Describe("Mkdir (and others)", func() {
		It("panics", func() {
			Expect(func() { fs.MkdirAll("test-path.json", os.ModeDir) }).To(Panic())
		})
	})
})
