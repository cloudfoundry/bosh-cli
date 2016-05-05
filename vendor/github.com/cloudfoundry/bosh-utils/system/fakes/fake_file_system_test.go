package fakes_test

import (
	"os"
	"path/filepath"

	. "github.com/cloudfoundry/bosh-utils/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-utils/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("FakeFileSystem", func() {
	var (
		fs *FakeFileSystem
	)

	BeforeEach(func() {
		fs = NewFakeFileSystem()
	})

	Describe("RemoveAll", func() {
		It("removes the specified file", func() {
			fs.WriteFileString("foobar", "asdfghjk")
			fs.WriteFileString("foobarbaz", "qwertyuio")

			err := fs.RemoveAll("foobar")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("foobar")).To(BeFalse())
			Expect(fs.FileExists("foobarbaz")).To(BeTrue())

			err = fs.RemoveAll("foobarbaz")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("foobarbaz")).To(BeFalse())
		})

		It("removes the specified dir and the files under it", func() {
			err := fs.MkdirAll("foobarbaz", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("foobarbaz/stuff.txt", "asdfghjk")
			Expect(err).ToNot(HaveOccurred())
			err = fs.MkdirAll("foobar", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("foobar/stuff.txt", "qwertyuio")
			Expect(err).ToNot(HaveOccurred())

			err = fs.RemoveAll("foobar")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("foobar")).To(BeFalse())
			Expect(fs.FileExists("foobar/stuff.txt")).To(BeFalse())
			Expect(fs.FileExists("foobarbaz")).To(BeTrue())
			Expect(fs.FileExists("foobarbaz/stuff.txt")).To(BeTrue())

			err = fs.RemoveAll("foobarbaz")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("foobarbaz")).To(BeFalse())
			Expect(fs.FileExists("foobarbaz/stuff.txt")).To(BeFalse())
		})

		It("removes the specified symlink (but not the file it links to)", func() {
			err := fs.WriteFileString("foobarbaz", "asdfghjk")
			Expect(err).ToNot(HaveOccurred())
			err = fs.Symlink("foobarbaz", "foobar")
			Expect(err).ToNot(HaveOccurred())

			err = fs.RemoveAll("foobarbaz")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("foobarbaz")).To(BeFalse())
			Expect(fs.FileExists("foobar")).To(BeTrue())

			err = fs.RemoveAll("foobar")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("foobar")).To(BeFalse())
		})
	})

	Describe("CopyDir", func() {
		var fixtureFiles = map[string]string{
			"foo.txt":         "asdfghjkl",
			"bar/bar.txt":     "qwertyuio",
			"bar/baz/bar.txt": "zxcvbnm,\nafawg",
		}

		var (
			fixtureDirPath = "fixtures"
		)

		BeforeEach(func() {
			for fixtureFile, contents := range fixtureFiles {
				fs.WriteFileString(filepath.Join(fixtureDirPath, fixtureFile), contents)
			}
		})

		It("recursively copies directory contents", func() {
			srcPath := fixtureDirPath
			dstPath, err := fs.TempDir("CopyDirTestDir")
			Expect(err).ToNot(HaveOccurred())
			defer fs.RemoveAll(dstPath)

			err = fs.CopyDir(srcPath, dstPath)
			Expect(err).ToNot(HaveOccurred())

			for fixtureFile := range fixtureFiles {
				srcContents, err := fs.ReadFile(filepath.Join(srcPath, fixtureFile))
				Expect(err).ToNot(HaveOccurred())

				dstContents, err := fs.ReadFile(filepath.Join(dstPath, fixtureFile))
				Expect(err).ToNot(HaveOccurred())

				Expect(srcContents).To(Equal(dstContents), "Copied file does not match source file: '%s", fixtureFile)
			}

			err = fs.RemoveAll(dstPath)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
