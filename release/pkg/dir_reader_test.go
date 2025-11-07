package pkg_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	fakeres "github.com/cloudfoundry/bosh-cli/v7/release/resource/resourcefakes"
)

var _ = Describe("DirReaderImpl", func() {
	var (
		srcDirPath = filepath.Join("/", "src")

		collectedFiles          []File
		collectedPrepFiles      []File
		collectedChunks         []string
		collectedFollowSymlinks bool
		collectedNoCompression  bool
		archive                 *fakeres.FakeArchive
		fs                      *fakesys.FakeFileSystem
		reader                  DirReaderImpl
	)

	BeforeEach(func() {
		archive = &fakeres.FakeArchive{}
		archiveFactory := func(args ArchiveFactoryArgs) Archive {
			collectedFiles = args.Files
			collectedPrepFiles = args.PrepFiles
			collectedChunks = args.Chunks
			collectedFollowSymlinks = args.FollowSymlinks
			collectedNoCompression = args.NoCompression
			return archive
		}
		fs = fakesys.NewFakeFileSystem()

		reader = NewDirReaderImpl(archiveFactory, srcDirPath, filepath.Join(srcDirPath, "blobs"), fs)
	})

	Describe("Read", func() {
		It("returns a package with the details collected from a directory", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
dependencies: [pkg1, pkg2]
files: [in-file1, in-file2]
excluded_files: [ex-file1, ex-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "in-file2"), []string{filepath.Join(srcDirPath, "in-file2")})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), []string{"pkg1", "pkg2"})))

			Expect(collectedFiles).To(ConsistOf(
				// does not include spec
				File{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
				File{Path: filepath.Join(srcDirPath, "in-file1"), DirPath: srcDirPath, RelativePath: "in-file1"},
				File{Path: filepath.Join(srcDirPath, "in-file2"), DirPath: srcDirPath, RelativePath: "in-file2"},
			))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(Equal([]string{"pkg1", "pkg2"}))
			Expect(collectedFollowSymlinks).To(BeFalse())
			Expect(collectedNoCompression).To(BeFalse())
		})

		It("creates archive factory with no_compression false when package spec does not contain no_compression", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
dependencies: [pkg1]
files: [in-file1]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})

			archive.FingerprintReturns("fp", nil)

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())

			Expect(collectedNoCompression).To(BeFalse())
		})

		It("creates archive factory with no_compression true when package spec contains no_compression: true", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
dependencies: [pkg1]
files: [in-file1]
no_compression: true
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})

			archive.FingerprintReturns("fp", nil)

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())

			Expect(collectedNoCompression).To(BeTrue())
		})

		It("returns a package with the details with pre_packaging file", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), "name: name")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "pre_packaging"), "")
			Expect(err).ToNot(HaveOccurred())

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "dir", "pre_packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "pre_packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedChunks).To(BeEmpty())
		})

		It("returns a package with src files and blob files", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
files: [in-file1, in-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "pre_packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "in-file2"), []string{filepath.Join(srcDirPath, "blobs", "in-file2")})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(ConsistOf([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "dir", "pre_packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "pre_packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "in-file1"), DirPath: srcDirPath, RelativePath: "in-file1"},
				{Path: filepath.Join(srcDirPath, "blobs", "in-file2"), DirPath: filepath.Join(srcDirPath, "blobs"), RelativePath: "in-file2"},
			}))

			Expect(collectedPrepFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "pre_packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedChunks).To(BeEmpty())
		})

		It("prefers src files over blob files", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
files: [in-file1, in-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "in-file2"), []string{filepath.Join(srcDirPath, "in-file2")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "in-file2"), []string{filepath.Join(srcDirPath, "blobs", "in-file2")})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(ConsistOf([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "in-file1"), DirPath: srcDirPath, RelativePath: "in-file1"},
				{Path: filepath.Join(srcDirPath, "in-file2"), DirPath: srcDirPath, RelativePath: "in-file2"},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("returns an error if glob doesnt match src or blob files", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
files: [in-file1, in-file2, missing-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "in-file2"), []string{filepath.Join(srcDirPath, "blobs", "in-file2")})

			// Directories are not packageable
			err = fs.MkdirAll(filepath.Join(srcDirPath, "missing-file2"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.MkdirAll(filepath.Join(srcDirPath, "blobs", "missing-file2"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "missing-file2"), []string{filepath.Join(srcDirPath, "missing-file2")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "missing-file2"), []string{filepath.Join(srcDirPath, "blobs", "missing-file2")})

			archive.FingerprintReturns("fp", nil)

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Missing files for pattern 'missing-file2'"))
		})

		It("excludes files and blobs", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
files: [in-file1, in-file2]
excluded_files: [ex-file1, ex-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "in-file1"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "in-file2"), []string{filepath.Join(srcDirPath, "blobs", "in-file2")})
			fs.SetGlob(filepath.Join(srcDirPath, "ex-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "ex-file2"), []string{filepath.Join(srcDirPath, "blobs", "in-file2")})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("allows to only have packaging file and to exclude all files", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
excluded_files: [ex-file1, ex-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
			}))

			Expect(collectedPrepFiles).To(BeEmpty())
			Expect(collectedChunks).To(BeEmpty())
		})

		It("matches files in blobs directory even if glob also matches empty folders in src directory", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
dependencies: [pkg1, pkg2]
files:
- "**/*"
excluded_files: [ex-file1, ex-file2]
`)
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "**", "*"), []string{filepath.Join(srcDirPath, "directory")})

			err = fs.MkdirAll(filepath.Join(srcDirPath, "directory"), 0777)
			Expect(err).NotTo(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).NotTo(HaveOccurred())

			err = fs.MkdirAll(filepath.Join(srcDirPath, "directory", "f1", "/"), 0777)
			Expect(err).NotTo(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "directory", "f1"), "")
			Expect(err).NotTo(HaveOccurred())

			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "**", "*"), []string{filepath.Join(srcDirPath, "blobs", "directory"), filepath.Join(srcDirPath, "blobs", "directory", "f1")})
			fs.SetGlob(filepath.Join(srcDirPath, "**", "*"), []string{filepath.Join(srcDirPath, "directory"), filepath.Join(srcDirPath, "directory", "f1")})

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())

			Expect(collectedFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "blobs", "directory", "f1"), DirPath: filepath.Join(srcDirPath, "blobs"), RelativePath: filepath.Join("directory", "f1")},
			}))
		})

		It("returns error if packaging is included in specified files", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), "name: name\nfiles: [packaging]")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "packaging"), []string{filepath.Join(srcDirPath, "packaging")})

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Expected special 'packaging' file to not be included via 'files' key for package 'name'"))
		})

		It("returns error if pre_packaging is included in specified files", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), "name: name\nfiles: [pre_packaging]")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "pre_packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "pre_packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			fs.SetGlob(filepath.Join(srcDirPath, "pre_packaging"), []string{filepath.Join(srcDirPath, "pre_packaging")})

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Expected special 'pre_packaging' file to not be included via 'files' key for package 'name'"))
		})

		It("returns error if spec file is not valid", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `-`)
			Expect(err).ToNot(HaveOccurred())

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Collecting package files"))
		})

		It("returns error if packaging file is not found", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), "name: name")
			Expect(err).ToNot(HaveOccurred())

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find '" + filepath.Join(srcDirPath, "dir", "packaging") + "' for package 'name'"))
		})

		globErrChecks := map[string]string{
			"src files dir (files)":            filepath.Join(srcDirPath, "file1"),
			"blobs files dir (files)":          filepath.Join(srcDirPath, "blobs", "file1"),
			"src files dir (excluded files)":   filepath.Join(srcDirPath, "ex-file1"),
			"blobs files dir (excluded files)": filepath.Join(srcDirPath, "blobs", "ex-file1"),
		}

		for desc, pattern := range globErrChecks {
			desc, pattern := desc, pattern // copy

			It(fmt.Sprintf("returns error if globbing '%s' (%s) fails", desc, pattern), func() {
				err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), "name: name\nfiles: [file1]\nexcluded_files: [ex-file1]")
				Expect(err).ToNot(HaveOccurred())
				err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
				Expect(err).ToNot(HaveOccurred())

				err = fs.WriteFileString(filepath.Join(srcDirPath, "file1"), "")
				Expect(err).ToNot(HaveOccurred())
				err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "file1"), "")
				Expect(err).ToNot(HaveOccurred())
				fs.SetGlob(filepath.Join(srcDirPath, "file1"), []string{filepath.Join(srcDirPath, "file1")})
				fs.SetGlob(filepath.Join(srcDirPath, "blobs", "file1"), []string{filepath.Join(srcDirPath, "blobs", "file1")})

				fs.GlobErrs[pattern] = errors.New("fake-err")

				_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		}

		It("returns error if fingerprinting fails", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())

			archive.FingerprintReturns("", errors.New("fake-err"))

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("include bad src symlinks", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
files: [in-file1, in-file2]
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "pre_packaging"), "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.Symlink(filepath.Join(srcDirPath, "invalid", "path"), filepath.Join(srcDirPath, "in-file1"))
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(srcDirPath, "blobs", "in-file2"), "")
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob(filepath.Join(srcDirPath, "in-file1"), []string{filepath.Join(srcDirPath, "in-file1")})
			fs.SetGlob(filepath.Join(srcDirPath, "blobs", "in-file2"), []string{filepath.Join(srcDirPath, "blobs", "in-file2")})

			archive.FingerprintReturns("fp", nil)

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

			Expect(collectedFiles).To(ConsistOf([]File{
				{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "dir", "pre_packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "pre_packaging", ExcludeMode: true},
				{Path: filepath.Join(srcDirPath, "in-file1"), DirPath: srcDirPath, RelativePath: "in-file1"},
				{Path: filepath.Join(srcDirPath, "blobs", "in-file2"), DirPath: filepath.Join(srcDirPath, "blobs"), RelativePath: "in-file2"},
			}))

			Expect(collectedPrepFiles).To(Equal([]File{
				{Path: filepath.Join(srcDirPath, "dir", "pre_packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "pre_packaging", ExcludeMode: true},
			}))

			Expect(collectedChunks).To(BeEmpty())
		})

		It("returns a package with spec lock", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec.lock"), "name: name\nfingerprint: fp\ndependencies: [pkg1]")
			Expect(err).ToNot(HaveOccurred())

			pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).ToNot(HaveOccurred())
			Expect(pkg).To(Equal(NewPackage(NewExistingResource("name", "fp", ""), []string{"pkg1"})))
		})

		It("returns error if cannot deserialize spec lock", func() {
			err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec.lock"), "-")
			Expect(err).ToNot(HaveOccurred())

			_, err = reader.Read(filepath.Join(srcDirPath, "dir"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unmarshalling package spec lock"))
		})

		Context("when the glob matches a path that contains a directory symlink", func() {
			BeforeEach(func() {
				err := fs.WriteFileString(filepath.Join(srcDirPath, "dir", "spec"), `---
name: name
files: ["stuff/**/*"]
`)
				Expect(err).ToNot(HaveOccurred())

				err = fs.WriteFileString(filepath.Join(srcDirPath, "dir", "packaging"), "")
				Expect(err).ToNot(HaveOccurred())
				err = fs.WriteFileString(filepath.Join(srcDirPath, "symlink-target", "file"), "")
				Expect(err).ToNot(HaveOccurred())
				err = fs.WriteFileString(filepath.Join(srcDirPath, "stuff", "symlink-dir", "file"), "")
				Expect(err).ToNot(HaveOccurred())

				err = fs.Symlink(
					filepath.Join(srcDirPath, "symlink-target"),
					filepath.Join(srcDirPath, "stuff", "symlink-dir"),
				)
				Expect(err).ToNot(HaveOccurred())

				fs.SetGlob(
					"/src/stuff/**/*",
					[]string{
						filepath.Join(srcDirPath, "stuff", "symlink-dir"),
						filepath.Join(srcDirPath, "stuff", "symlink-dir", "file"),
					},
				)
				archive.FingerprintReturns("fp", nil)
			})

			It("does not include any files that are nested underneath the symlink", func() {
				pkg, err := reader.Read(filepath.Join(srcDirPath, "dir"))
				Expect(err).NotTo(HaveOccurred())
				Expect(pkg).To(Equal(NewPackage(NewResource("name", "fp", archive), nil)))

				Expect(collectedFiles).To(ConsistOf([]File{
					{Path: filepath.Join(srcDirPath, "dir", "packaging"), DirPath: filepath.Join(srcDirPath, "dir"), RelativePath: "packaging", ExcludeMode: true},
					{Path: filepath.Join(srcDirPath, "stuff", "symlink-dir"), DirPath: srcDirPath, RelativePath: "stuff/symlink-dir", ExcludeMode: false},
				}))
			})
		})
	})
})
