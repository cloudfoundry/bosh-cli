package pkg_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/release/pkg"
	. "github.com/cloudfoundry/bosh-init/release/resource"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeres "github.com/cloudfoundry/bosh-init/release/resource/fakes"
)

var _ = Describe("DirReaderImpl", func() {
	var (
		collectedFiles     []File
		collectedPrepFiles []File
		collectedChunks    []string
		archive            *fakeres.FakeArchive
		fs                 *fakesys.FakeFileSystem
		reader             DirReaderImpl
	)

	BeforeEach(func() {
		archive = &fakeres.FakeArchive{}
		archiveFactory := func(files, prepFiles []File, chunks []string) Archive {
			collectedFiles = files
			collectedPrepFiles = prepFiles
			collectedChunks = chunks
			return archive
		}
		fs = fakesys.NewFakeFileSystem()

		reader = NewDirReaderImpl(archiveFactory, "/src", "/blobs", fs)
	})

	Describe("Read", func() {

		Context("when packaging path contains folders", func() {
			var err error

			BeforeEach(func(){
				fs.WriteFileString("/dir/spec", `---
name: name
dependencies: [pkg1, pkg2]
files:
- "**/*"
excluded_files: [ex-file1, ex-file2]
`)
				fs.SetGlob("/src/**/*", []string{"/src/directory"})
				err = fs.MkdirAll("/src/directory", 0777)
				Expect(err).NotTo(HaveOccurred())

				err = fs.WriteFileString("/dir/packaging", "")
				Expect(err).NotTo(HaveOccurred())
			})

			It("ignores it", func(){
				_, err = reader.Read("/dir")
				Expect(err).NotTo(HaveOccurred())


				Expect(collectedFiles).To(Equal([]File{
					File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
				}))
			})
		})

		It("returns a package with the details collected from a directory", func() {
			fs.WriteFileString("/dir/spec", `---
name: name
dependencies: [pkg1, pkg2]
files: [in-file1, in-file2]
excluded_files: [ex-file1, ex-file2]
`)

			fs.WriteFileString("/dir/packaging", "")

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read("/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), []string{"pkg1", "pkg2"})))

			Expect(collectedFiles).To(Equal([]File{
				// does not include spec
				File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(Equal([]string{"pkg1", "pkg2"}))
		})

		It("returns a package with the details with pre_packaging file", func() {
			fs.WriteFileString("/dir/spec", "name: name")
			fs.WriteFileString("/dir/packaging", "")
			fs.WriteFileString("/dir/pre_packaging", "")

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read("/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(Equal([]File{
				File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
				File{Path: "/dir/pre_packaging", DirPath: "/dir", RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(Equal([]File{
				File{Path: "/dir/pre_packaging", DirPath: "/dir", RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedChunks).To(BeEmpty())
		})

		It("returns a package with src files and blob files", func() {
			fs.WriteFileString("/dir/spec", `---
name: name
files: [in-file1, in-file2]
`)

			fs.WriteFileString("/dir/packaging", "")
			fs.WriteFileString("/dir/pre_packaging", "")
			fs.WriteFileString("/src/in-file1", "")
			fs.WriteFileString("/blobs/in-file2", "")

			fs.SetGlob("/src/in-file1", []string{"/src/in-file1"})
			fs.SetGlob("/blobs/in-file2", []string{"/blobs/in-file2"})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read("/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(ConsistOf([]File{
				File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
				File{Path: "/dir/pre_packaging", DirPath: "/dir", RelativePath: "pre_packaging", ExcludeMode: true},
				File{Path: "/src/in-file1", DirPath: "/src", RelativePath: "in-file1"},
				File{Path: "/blobs/in-file2", DirPath: "/blobs", RelativePath: "in-file2"},
			}))

			Expect(collectedPrepFiles).To(Equal([]File{
				File{Path: "/dir/pre_packaging", DirPath: "/dir", RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedChunks).To(BeEmpty())
		})

		It("prefers src files over blob files", func() {
			fs.WriteFileString("/dir/spec", `---
name: name
files: [in-file1, in-file2]
`)

			fs.WriteFileString("/dir/packaging", "")
			fs.WriteFileString("/src/in-file1", "")
			fs.WriteFileString("/src/in-file2", "")
			fs.WriteFileString("/blobs/in-file2", "")

			fs.SetGlob("/src/in-file1", []string{"/src/in-file1"})
			fs.SetGlob("/src/in-file2", []string{"/src/in-file2"})
			fs.SetGlob("/blobs/in-file2", []string{"/blobs/in-file2"})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read("/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(ConsistOf([]File{
				File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
				File{Path: "/src/in-file1", DirPath: "/src", RelativePath: "in-file1"},
				File{Path: "/src/in-file2", DirPath: "/src", RelativePath: "in-file2"},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("excludes files and blobs", func() {
			fs.WriteFileString("/dir/spec", `---
name: name
files: [in-file1, in-file2]
excluded_files: [ex-file1, ex-file2]
`)

			fs.WriteFileString("/dir/packaging", "")
			fs.WriteFileString("/src/in-file1", "")
			fs.WriteFileString("/blobs/in-file2", "")

			fs.SetGlob("/src/in-file1", []string{"/src/in-file1"})
			fs.SetGlob("/blobs/in-file2", []string{"/blobs/in-file2"})
			fs.SetGlob("/src/ex-file1", []string{"/src/in-file1"})
			fs.SetGlob("/blobs/ex-file2", []string{"/blobs/in-file2"})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read("/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(Equal([]File{
				File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("allows to only have packaging file and to exclude all files", func() {
			fs.WriteFileString("/dir/spec", `---
name: name
excluded_files: [ex-file1, ex-file2]
`)

			fs.WriteFileString("/dir/packaging", "")

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read("/dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(Equal([]File{
				File{Path: "/dir/packaging", DirPath: "/dir", RelativePath: "packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("returns error if packaging is included in specified files", func() {
			fs.WriteFileString("/dir/spec", "name: name\nfiles: [packaging]")

			fs.WriteFileString("/dir/packaging", "")
			fs.WriteFileString("/src/packaging", "")
			fs.SetGlob("/src/packaging", []string{"/src/packaging"})

			_, err := reader.Read("/dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Expected special 'packaging' file to not be included via 'files' key for package 'name'"))
		})

		It("returns error if pre_packaging is included in specified files", func() {
			fs.WriteFileString("/dir/spec", "name: name\nfiles: [pre_packaging]")

			fs.WriteFileString("/dir/packaging", "")
			fs.WriteFileString("/dir/pre_packaging", "")
			fs.WriteFileString("/src/pre_packaging", "")
			fs.SetGlob("/src/pre_packaging", []string{"/src/pre_packaging"})

			_, err := reader.Read("/dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Expected special 'pre_packaging' file to not be included via 'files' key for package 'name'"))
		})

		It("returns error if spec file is not valid", func() {
			fs.WriteFileString("/dir/spec", `-`)

			_, err := reader.Read("/dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Collecting package files"))
		})

		It("returns error if packaging file is not found", func() {
			fs.WriteFileString("/dir/spec", "name: name")

			_, err := reader.Read("/dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find '/dir/packaging' for package 'name'"))
		})

		globErrChecks := map[string]string{
			"src files dir (files)":            "/src/file1",
			"blobs files dir (files)":          "/blobs/file1",
			"src files dir (excluded files)":   "/src/ex-file1",
			"blobs files dir (excluded files)": "/blobs/ex-file1",
		}

		for desc, pattern := range globErrChecks {
			It(fmt.Sprintf("returns error if globbing %s fails", desc), func() {
				fs.WriteFileString("/dir/spec", "name: name\nfiles: [file1]\nexcluded_files: [ex-file1]")
				fs.WriteFileString("/dir/packaging", "")

				fs.GlobErrs[pattern] = errors.New("fake-err")

				_, err := reader.Read("/dir")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		}

		It("returns error if fingerprinting fails", func() {
			fs.WriteFileString("/dir/spec", "")
			fs.WriteFileString("/dir/packaging", "")

			archive.FingerprintReturns("", errors.New("fake-err"))

			_, err := reader.Read("/dir")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
