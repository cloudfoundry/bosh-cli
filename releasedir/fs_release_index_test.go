package releasedir_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
	fakerel "github.com/cloudfoundry/bosh-cli/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/releasedir/releasedirfakes"
)

var _ = Describe("FSReleaseIndex", func() {
	var (
		reporter *fakereldir.FakeReleaseIndexReporter
		uuidGen  *fakeuuid.FakeGenerator
		fs       *fakesys.FakeFileSystem
		index    FSReleaseIndex
	)

	BeforeEach(func() {
		reporter = &fakereldir.FakeReleaseIndexReporter{}
		uuidGen = &fakeuuid.FakeGenerator{}
		fs = fakesys.NewFakeFileSystem()
		index = NewFSReleaseIndex("index-name", "/dir", reporter, uuidGen, fs)
	})

	Describe("LastVersion", func() {
		It("returns nil when there is no index file", func() {
			ver, err := index.LastVersion("name")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver).To(BeNil())
		})

		It("returns nil when index file is empty", func() {
			fs.WriteFileString("/dir/name/index.yml", "")

			ver, err := index.LastVersion("name")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver).To(BeNil())
		})

		It("returns greater version", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  uuid1: {version: "1.1"}
  uuid2: {version: "1"}
format-version: "2"`)

			ver, err := index.LastVersion("name")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1").String()))
		})

		It("returns error if version cannot be parsed", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  uuid2: {version: "-"}
format-version: "2"`)

			_, err := index.LastVersion("name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Parsing release versions"))
		})

		It("returns error if name is empty", func() {
			_, err := index.LastVersion("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty release name"))
		})

		It("returns error if index file cannot be unmarshalled", func() {
			fs.WriteFileString("/dir/name/index.yml", "-")

			_, err := index.LastVersion("name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if reading index file fails", func() {
			fs.WriteFileString("/dir/name/index.yml", "")
			fs.ReadFileError = errors.New("fake-err")

			_, err := index.LastVersion("name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Contains", func() {
		var (
			release *fakerel.FakeRelease
		)

		BeforeEach(func() {
			release = &fakerel.FakeRelease{}
			release.NameReturns("name")
			release.VersionReturns("ver1")
		})

		It("returns false when there is no index file", func() {
			exists, err := index.Contains(release)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("returns false when index file is empty", func() {
			fs.WriteFileString("/dir/name/index.yml", "")

			exists, err := index.Contains(release)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("returns true if version is exists", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  uuid1: {version: "1.1"}
  uuid2: {version: "ver1"}
format-version: "2"`)

			exists, err := index.Contains(release)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("returns false if version is not exists", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  uuid1: {version: "1.1"}
format-version: "2"`)

			exists, err := index.Contains(release)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("returns error if name is empty", func() {
			release.NameReturns("")

			_, err := index.Contains(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty release name"))
		})

		It("returns error if version is empty", func() {
			release.VersionReturns("")

			_, err := index.Contains(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty release version"))
		})

		It("returns error if index file cannot be unmarshalled", func() {
			fs.WriteFileString("/dir/name/index.yml", "-")

			_, err := index.Contains(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if reading index file fails", func() {
			fs.WriteFileString("/dir/name/index.yml", "")
			fs.ReadFileError = errors.New("fake-err")

			_, err := index.Contains(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Add", func() {
		var (
			manifest boshman.Manifest
		)

		BeforeEach(func() {
			manifest.Name = "name"
			manifest.Version = "ver1"
			uuidGen.GeneratedUUID = "new-uuid"
		})

		It("saves manifest and adds version entry when there is no index file", func() {
			err := index.Add(manifest)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/name/name-ver1.yml")).To(Equal(`name: name
version: ver1
commit_hash: ""
uncommitted_changes: false
`))

			Expect(fs.ReadFileString("/dir/name/index.yml")).To(Equal(`builds:
  new-uuid:
    version: ver1
format-version: "2"
`))

			name, desc, err := reporter.ReleaseIndexAddedArgsForCall(0)
			Expect(name).To(Equal("index-name"))
			Expect(desc).To(Equal("name/ver1"))
			Expect(err).To(BeNil())
		})

		It("saves manifest and adds version entry", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  uuid: {version: "1.1"}
format-version: "2"
`)

			err := index.Add(manifest)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/name/name-ver1.yml")).To(Equal(`name: name
version: ver1
commit_hash: ""
uncommitted_changes: false
`))

			Expect(fs.ReadFileString("/dir/name/index.yml")).To(Equal(`builds:
  new-uuid:
    version: ver1
  uuid:
    version: "1.1"
format-version: "2"
`))

			name, desc, err := reporter.ReleaseIndexAddedArgsForCall(0)
			Expect(name).To(Equal("index-name"))
			Expect(desc).To(Equal("name/ver1"))
			Expect(err).To(BeNil())
		})

		It("returns and reports error if writing manifest fails", func() {
			fs.WriteFileErrors["/dir/name/name-ver1.yml"] = errors.New("fake-err")

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())

			name, desc, err := reporter.ReleaseIndexAddedArgsForCall(0)
			Expect(name).To(Equal("index-name"))
			Expect(desc).To(Equal("name/ver1"))
			Expect(err).ToNot(BeNil())
		})

		It("returns and reports error if writing index fails", func() {
			fs.WriteFileErrors["/dir/name/index.yml"] = errors.New("fake-err")

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())

			name, desc, err := reporter.ReleaseIndexAddedArgsForCall(0)
			Expect(name).To(Equal("index-name"))
			Expect(desc).To(Equal("name/ver1"))
			Expect(err).ToNot(BeNil())
		})

		It("returns error if generating uuid fails", func() {
			uuidGen.GenerateError = errors.New("fake-err")

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if version is exists", func() {
			fs.WriteFileString("/dir/name/index.yml", `---
builds:
  uuid1: {version: "1.1"}
  uuid2: {version: "ver1"}
format-version: "2"`)

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Release version 'ver1' already exists"))
		})

		It("returns error if name is empty", func() {
			manifest.Name = ""

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty release name"))
		})

		It("returns error if version is empty", func() {
			manifest.Version = ""

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty release version"))
		})

		It("returns error if index file cannot be unmarshalled", func() {
			fs.WriteFileString("/dir/name/index.yml", "-")

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if reading index file fails", func() {
			fs.WriteFileString("/dir/name/index.yml", "")
			fs.ReadFileError = errors.New("fake-err")

			err := index.Add(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("ManifestPath", func() {
		It("returns path to a manifest", func() {
			Expect(index.ManifestPath("name", "ver1")).To(Equal("/dir/name/name-ver1.yml"))
		})
	})
})
