package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cloudfoundry/bosh-utils/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshrelman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("create-release command", func() {
	var (
		ui         *fakeui.FakeUI
		fs         boshsys.FileSystem
		deps       BasicDeps
		cmdFactory Factory
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		confUI := boshui.NewWrappingConfUI(ui, logger)

		fs = boshsys.NewOsFileSystem(logger)
		deps = NewBasicDepsWithFS(confUI, fs, logger)
		cmdFactory = NewFactory(deps)
	})

	removeSHA256s := func(contents string) string {
		matchSHA256s := regexp.MustCompile("sha1: sha256:[a-z0-9]{64}\n")
		return matchSHA256s.ReplaceAllString(contents, "sha1: replaced\n")
	}

	expectSha256Checksums := func(filePath string) {
		contents, err := fs.ReadFileString(filePath)
		Expect(err).ToNot(HaveOccurred())
		Expect(contents).To(MatchRegexp("sha1: sha256:.*"))
	}

	It("can iterate on a basic release", func() {
		suffix, err := uuid.NewGenerator().Generate()
		Expect(err).ToNot(HaveOccurred())

		// containing the release in a directory that is a symlink
		// to ensure we can work inside symlinks (i.e. macOS /tmp)
		containerDir := filepath.Join("/", "tmp", suffix)
		symlinkedContainerDir := fmt.Sprintf("%s-symlinked", containerDir)
		err = fs.MkdirAll(containerDir, 0755)
		Expect(err).ToNot(HaveOccurred())
		err = fs.Symlink(containerDir, symlinkedContainerDir)
		Expect(err).ToNot(HaveOccurred())
		tmpDir := filepath.Join(symlinkedContainerDir, "release")

		defer func() {
			err = fs.RemoveAll(containerDir)
			Expect(err).ToNot(HaveOccurred())
			err = fs.RemoveAll(symlinkedContainerDir)
			Expect(err).ToNot(HaveOccurred())
		}()

		relName := filepath.Base(tmpDir)

		{
			execCmd(cmdFactory, []string{"init-release", "--dir", tmpDir})
			Expect(fs.FileExists(filepath.Join(tmpDir, "config"))).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(tmpDir, "jobs"))).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(tmpDir, "packages"))).To(BeTrue())
			Expect(fs.FileExists(filepath.Join(tmpDir, "src"))).To(BeTrue())
		}

		execCmd(cmdFactory, []string{"generate-job", "job1", "--dir", tmpDir})
		execCmd(cmdFactory, []string{"generate-package", "pkg1", "--dir", tmpDir})
		execCmd(cmdFactory, []string{"generate-package", "pkg2", "--dir", tmpDir})

		err = fs.WriteFileString(filepath.Join(tmpDir, "LICENSE"), "LICENSE")
		Expect(err).ToNot(HaveOccurred())

		{ // pkg1 depends on pkg2 for compilation
			pkg1SpecPath := filepath.Join(tmpDir, "packages", "pkg1", "spec")

			contents, err := fs.ReadFileString(pkg1SpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "dependencies: []", "dependencies: [pkg2]", -1))
			Expect(err).ToNot(HaveOccurred())
		}

		{ // job1 depends on both packages
			jobSpecPath := filepath.Join(tmpDir, "jobs", "job1", "spec")

			contents, err := fs.ReadFileString(jobSpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1, pkg2]", -1))
			Expect(err).ToNot(HaveOccurred())
		}

		{ // Make empty release
			execCmd(cmdFactory, []string{"create-release", "--dir", tmpDir})

			contents, err := fs.ReadFileString(filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.1.yml"))
			Expect(err).ToNot(HaveOccurred())

			Expect(removeSHA256s(contents)).To(Equal(
				"name: " + relName + `
version: 0+dev.1
commit_hash: non-git
uncommitted_changes: false
jobs:
- name: job1
  version: f54520d6563c438bf0bc5bb674777db171b78d848a057a3faec0e9b572c3a76c
  fingerprint: f54520d6563c438bf0bc5bb674777db171b78d848a057a3faec0e9b572c3a76c
  sha1: replaced
  packages:
  - pkg1
  - pkg2
packages:
- name: pkg1
  version: 08441a1962e8141645edb0f2ddb91330454f2f1a3954d7f27fa256eb5e7b4ed6
  fingerprint: 08441a1962e8141645edb0f2ddb91330454f2f1a3954d7f27fa256eb5e7b4ed6
  sha1: replaced
  dependencies:
  - pkg2
- name: pkg2
  version: 34581dd0d93735e444a32450e3ae3951258c936479b45e08f1fa074740c7e392
  fingerprint: 34581dd0d93735e444a32450e3ae3951258c936479b45e08f1fa074740c7e392
  sha1: replaced
  dependencies: []
license:
  version: 42a33a7295936a632c8f54e70f2553975ee38a476d6aae93f3676e68c9db2f86
  fingerprint: 42a33a7295936a632c8f54e70f2553975ee38a476d6aae93f3676e68c9db2f86
  sha1: replaced
`))
		}

		{ // Add a bit of content
			err := fs.WriteFileString(filepath.Join(tmpDir, "src", "in-src"), "in-src")
			Expect(err).ToNot(HaveOccurred())

			randomFile := filepath.Join(tmpDir, "random-file")

			err = fs.WriteFileString(randomFile, "in-blobs")
			Expect(err).ToNot(HaveOccurred())

			execCmd(cmdFactory, []string{"add-blob", randomFile, "in-blobs", "--dir", tmpDir})

			pkg1SpecPath := filepath.Join(tmpDir, "packages", "pkg1", "spec")

			contents, err := fs.ReadFileString(pkg1SpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "files: []", "files:\n- in-src\n- in-blobs", -1))
			Expect(err).ToNot(HaveOccurred())
		}

		{ // Make release with some contents
			execCmd(cmdFactory, []string{"create-release", "--dir", tmpDir})

			rel1File := filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.1.yml")
			rel2File := filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.2.yml")

			contents, err := fs.ReadFileString(rel2File)
			Expect(err).ToNot(HaveOccurred())

			Expect(removeSHA256s(contents)).To(Equal(
				"name: " + relName + `
version: 0+dev.2
commit_hash: non-git
uncommitted_changes: false
jobs:
- name: job1
  version: f54520d6563c438bf0bc5bb674777db171b78d848a057a3faec0e9b572c3a76c
  fingerprint: f54520d6563c438bf0bc5bb674777db171b78d848a057a3faec0e9b572c3a76c
  sha1: replaced
  packages:
  - pkg1
  - pkg2
packages:
- name: pkg1
  version: 00ebebd8dd5a533a91f9de34b0cf708772fca87ada7e37e63bec00ece2e0634c
  fingerprint: 00ebebd8dd5a533a91f9de34b0cf708772fca87ada7e37e63bec00ece2e0634c
  sha1: replaced
  dependencies:
  - pkg2
- name: pkg2
  version: 34581dd0d93735e444a32450e3ae3951258c936479b45e08f1fa074740c7e392
  fingerprint: 34581dd0d93735e444a32450e3ae3951258c936479b45e08f1fa074740c7e392
  sha1: replaced
  dependencies: []
license:
  version: 42a33a7295936a632c8f54e70f2553975ee38a476d6aae93f3676e68c9db2f86
  fingerprint: 42a33a7295936a632c8f54e70f2553975ee38a476d6aae93f3676e68c9db2f86
  sha1: replaced
`,
			))

			man1, err := boshrelman.NewManifestFromPath(rel1File, fs)
			Expect(err).ToNot(HaveOccurred())

			man2, err := boshrelman.NewManifestFromPath(rel2File, fs)
			Expect(err).ToNot(HaveOccurred())

			// Explicitly check that pkg1 changed its fingerprint
			Expect(man1.Packages[0].Name).To(Equal(man2.Packages[0].Name))
			Expect(man1.Packages[0].Fingerprint).ToNot(Equal(man2.Packages[0].Fingerprint))

			// and pkg2 did not change
			Expect(man1.Packages[1].Name).To(Equal(man2.Packages[1].Name))
			Expect(man1.Packages[1].Fingerprint).To(Equal(man2.Packages[1].Fingerprint))
		}

		{ // check contents of index files when sha2 flag is supplied
			execCmd(cmdFactory, []string{"create-release", "--sha2", "--dir", tmpDir})

			expectSha256Checksums(filepath.Join(tmpDir, "dev_releases", relName, relName+"-0+dev.3.yml"))
			expectSha256Checksums(filepath.Join(tmpDir, ".dev_builds", "jobs", "job1", "index.yml"))
			expectSha256Checksums(filepath.Join(tmpDir, ".dev_builds", "packages", "pkg1", "index.yml"))
			expectSha256Checksums(filepath.Join(tmpDir, ".dev_builds", "license", "index.yml"))
		}

		{ // Check contents of made release via its tarball
			tgzFile := filepath.Join(tmpDir, "release-3.tgz")

			execCmd(cmdFactory, []string{"create-release", "--dir", tmpDir, "--tarball", tgzFile})
			relProvider := boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
			extractingArchiveReader := relProvider.NewExtractingArchiveReader()

			extractingRelease, err := extractingArchiveReader.Read(tgzFile)
			Expect(err).ToNot(HaveOccurred())

			defer extractingRelease.CleanUp() //nolint:errcheck

			pkg1 := extractingRelease.Packages()[0]
			Expect(fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-src"))).To(Equal("in-src"))
			Expect(fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-blobs"))).To(Equal("in-blobs"))

			archiveReader := relProvider.NewArchiveReader()

			release, err := archiveReader.Read(tgzFile)
			Expect(err).ToNot(HaveOccurred())

			defer release.CleanUp() //nolint:errcheck

			job1 := release.Jobs()[0]
			Expect(job1.PackageNames).To(ConsistOf("pkg1", "pkg2"))
		}

		{ // Check that tarballs will not overwrite a directory
			directoryPath := filepath.Join(tmpDir, "tarball-collision-dir")
			Expect(fs.MkdirAll(directoryPath, os.ModeDir)).To(Succeed())
			_, err := cmdFactory.New([]string{"create-release", "--dir", tmpDir, "--tarball", directoryPath})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Path must not be directory"))
		}

		{ // removes unknown blobs, keeping known blobs
			blobPath := filepath.Join(tmpDir, "blobs", "unknown-blob.tgz")

			err := fs.WriteFileString(blobPath, "i don't belong here")
			Expect(err).ToNot(HaveOccurred())

			execCmd(cmdFactory, []string{"create-release", "--dir", tmpDir})
			Expect(fs.FileExists(blobPath)).To(BeFalse())
			Expect(fs.FileExists(filepath.Join(tmpDir, "blobs", "in-blobs"))).To(BeTrue())
		}
	})
})
