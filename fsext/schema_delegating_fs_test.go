package fsext_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/fsext"
)

var _ = Describe("SchemaDelegatingFS", func() {
	var (
		fs         *SchemaDelegatingFS
		defaultFS  *fakesys.FakeFileSystem
		initTestFS *fakesys.FakeFileSystem
		testFS     *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		defaultFS = fakesys.NewFakeFileSystem()

		testFS = fakesys.NewFakeFileSystem()
		initTestFS = fakesys.NewFakeFileSystem()

		fs = NewSchemaDelegatingFS(defaultFS)
		fs.RegisterSchema("test-schema", testFS)
	})

	Describe("ExpandPath", func() {
		// Example configurations:
		// 	 /state.json                 ""
		// 	 state.json                  ""
		// 	 C:/state.json               ""
		// 	 file://state.json           file
		// 	 config-server://state.json  config-server

		It("returns expanded path by found fs", func() {
			testFS.ExpandPathExpanded = "expanded-path"

			path, err := fs.ExpandPath("test-schema://test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("test-schema://expanded-path"))

			testFS.ExpandPathExpanded = "test://path.json"

			path, err = fs.ExpandPath("test-schema://test://path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("test-schema://test://path.json"))
		})

		It("returns expanded path by default fs", func() {
			defaultFS.ExpandPathExpanded = "expanded-path"

			path, err := fs.ExpandPath("test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("expanded-path"))

			path, err = fs.ExpandPath("/test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("expanded-path"))

			path, err = fs.ExpandPath("C:/test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("expanded-path"))

			path, err = fs.ExpandPath("C:\\test-path.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("expanded-path"))
		})

		It("returns error if expanding path fails", func() {
			defaultFS.ExpandPathErr = errors.New("fake-err")

			_, err := fs.ExpandPath("test-path.json")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("fake-err"))
		})

		It("panic if unknown schema is given", func() {
			Expect(func() { fs.ExpandPath("unknown-schema://test-path.json") }).To(Panic())
		})
	})
})
