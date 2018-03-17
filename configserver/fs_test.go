package configserver_test

import (
	"os"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/configserver"
)

var _ = Describe("FS", func() {
	var (
		client *MemoryClient
		fs     boshsys.FileSystem
	)

	BeforeEach(func() {
		client = NewMemoryClient()
		fs = NewFS(client)
	})

	Describe("ExpandPath", func() {
		It("returns original path", func() {
			path, err := fs.ExpandPath("test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("test-path.json"))
		})
	})

	Describe("ReadFileWithOpts", func() {
		It("uses config server client", func() {
			client.Write("test-path-json", "test-content")

			content, err := fs.ReadFileWithOpts("test-path.json", boshsys.ReadOpts{})
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal([]byte("test-content")))
		})

		It("returns error if config server client errors", func() {
			_, err := fs.ReadFileWithOpts("test-path.json", boshsys.ReadOpts{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find 'test-path-json'"))
		})
	})

	Describe("FileExists", func() {
		It("uses config server client", func() {
			client.Write("test-path-json", "test-content")

			found := fs.FileExists("test-path.json")
			Expect(found).To(BeTrue())

			found = fs.FileExists("test-path.json-not-there")
			Expect(found).To(BeFalse())
		})
	})

	Describe("WriteFile", func() {
		It("uses config server client", func() {
			err := fs.WriteFile("test-path.json", []byte("test-content"))
			Expect(err).ToNot(HaveOccurred())

			Expect(client.Read("test-path-json")).To(Equal([]byte("test-content")))
			Expect(client.Exists("test-path-json")).To(BeTrue())
		})
	})

	Describe("RemoveAll", func() {
		It("uses config server client", func() {
			err := fs.WriteFile("test-path.json", []byte("test-content"))
			Expect(err).ToNot(HaveOccurred())

			Expect(client.Exists("test-path-json")).To(BeTrue())

			err = fs.RemoveAll("test-path.json")
			Expect(err).ToNot(HaveOccurred())

			Expect(client.Exists("test-path-json")).To(BeFalse())
		})

		It("does not fail if key does not exist", func() {
			err := fs.RemoveAll("test-path-does-not-exist.json")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("MkdirAll (and others)", func() {
		It("panics", func() {
			Expect(func() { fs.MkdirAll("test-path.json", os.ModeDir) }).To(Panic())
		})
	})
})
