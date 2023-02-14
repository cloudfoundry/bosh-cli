package releasedir_test

import (
	"errors"
	"os"
	"path/filepath"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

var _ = Describe("FSGenerator", func() {
	var (
		fs  *fakesys.FakeFileSystem
		gen FSGenerator
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		gen = NewFSGenerator("/dir", fs)
	})

	Describe("GenerateJob", func() {
		It("makes job directory", func() {
			err := fs.WriteFileString(filepath.Join("/", "dir", "public.yml"), "name: name")
			Expect(err).ToNot(HaveOccurred())

			err = gen.GenerateJob("job1")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "spec"))).To(Equal(`---
name: job1

templates: {}

packages: []

properties: {}
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "monit"))).To(Equal(""))

			Expect(fs.FileExists(filepath.Join("/", "dir", "jobs", "job1", "templates"))).To(BeTrue())
		})

		It("returns error if job directory already exists", func() {
			err := fs.MkdirAll(filepath.Join("/", "dir", "jobs", "job1"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = gen.GenerateJob("job1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Job 'job1' at '" + filepath.Join("/", "dir", "jobs", "job1") + "' already exists"))
		})

		It("returns error if fails to create job directory", func() {
			fs.MkdirAllError = errors.New("fake-err")

			err := gen.GenerateJob("job1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("GeneratePackage", func() {
		It("makes package directory", func() {
			err := fs.WriteFileString(filepath.Join("/", "dir", "public.yml"), "name: name")
			Expect(err).ToNot(HaveOccurred())

			err = gen.GeneratePackage("pkg1")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "packages", "pkg1", "spec"))).To(Equal(`---
name: pkg1

dependencies: []

files: []
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "packages", "pkg1", "packaging"))).To(Equal("set -e\n"))
		})

		It("returns error if package directory already exists", func() {
			err := fs.MkdirAll(filepath.Join("/", "dir", "packages", "pkg1"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = gen.GeneratePackage("pkg1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Package 'pkg1' at '" + filepath.Join("/", "dir", "packages", "pkg1") + "' already exists"))
		})

		It("returns error if fails to create package directory", func() {
			fs.MkdirAllError = errors.New("fake-err")

			err := gen.GeneratePackage("pkg1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
