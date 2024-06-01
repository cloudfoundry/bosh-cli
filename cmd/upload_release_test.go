package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("UploadReleaseCmd", func() {
	var (
		releaseReader *fakerel.FakeReader
		releaseWriter *fakerel.FakeWriter
		releaseDir    *fakereldir.FakeReleaseDir
		director      *fakedir.FakeDirector
		cmdRunner     *fakesys.FakeCmdRunner
		fs            *fakesys.FakeFileSystem
		archive       *fakedir.FakeReleaseArchive
		ui            *fakeui.FakeUI
		command       cmd.UploadReleaseCmd
	)

	BeforeEach(func() {
		releaseReader = &fakerel.FakeReader{}
		releaseDir = &fakereldir.FakeReleaseDir{}

		releaseDirFactory := func(dir opts.DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir) {
			Expect(dir).To(Equal(opts.DirOrCWDArg{Path: "/dir"}))
			return releaseReader, releaseDir
		}

		releaseWriter = &fakerel.FakeWriter{}
		releaseWriter.WriteReturns("/archive-path", nil)

		director = &fakedir.FakeDirector{}
		cmdRunner = fakesys.NewFakeCmdRunner()
		fs = fakesys.NewFakeFileSystem()

		archive = &fakedir.FakeReleaseArchive{}

		releaseArchiveFactory := func(path string) boshdir.ReleaseArchive {
			if archive.FileStub == nil {
				archive.FileStub = func() (boshdir.UploadFile, error) {
					return fakesys.NewFakeFile(path, fs), nil
				}
			}
			return archive
		}

		ui = &fakeui.FakeUI{}

		command = cmd.NewUploadReleaseCmd(releaseDirFactory, releaseWriter, director, releaseArchiveFactory, cmdRunner, fs, ui)
	})

	Describe("Run", func() {
		var (
			uploadReleaseOpts opts.UploadReleaseOpts
		)

		BeforeEach(func() {
			uploadReleaseOpts = opts.UploadReleaseOpts{
				Directory: opts.DirOrCWDArg{Path: "/dir"},
			}
		})

		act := func() error { return command.Run(uploadReleaseOpts) }

		Context("when url is remote (http/https)", func() {
			BeforeEach(func() {
				uploadReleaseOpts.Args.URL = "https://some-file.tzg"
			})

			It("uploads given release", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(1))

				url, sha1, rebase, fix := director.UploadReleaseURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal(""))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeFalse())
			})

			It("uploads given release even if reader is nil", func() {
				command = cmd.NewUploadReleaseCmd(nil, nil, director, nil, nil, nil, ui)

				err := command.Run(uploadReleaseOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(1))
			})

			It("uploads given release with a fix flag without checking if release exists", func() {
				uploadReleaseOpts.Fix = true

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.HasReleaseCallCount()).To(Equal(0))

				Expect(director.UploadReleaseURLCallCount()).To(Equal(1))

				url, sha1, rebase, fix := director.UploadReleaseURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal(""))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeTrue())
			})

			It("uploads given release with a specified rebase, sha1, etc.", func() {
				uploadReleaseOpts.Rebase = true
				uploadReleaseOpts.SHA1 = "sha1"

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(1))

				url, sha1, rebase, fix := director.UploadReleaseURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal("sha1"))
				Expect(rebase).To(BeTrue())
				Expect(fix).To(BeFalse())
			})

			It("does not upload release if name and version match existing release", func() {
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("existing-ver"))

				director.HasReleaseReturns(true, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(0))

				name, version, stemcell := director.HasReleaseArgsForCall(0)
				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("existing-ver"))
				Expect(stemcell).To(Equal(boshdir.OSVersionSlug{}))

				Expect(ui.Said).To(Equal(
					[]string{"Release 'existing-name/existing-ver' already exists."}))
			})

			It("does not upload compiled release if name, version and stemcell match existing release", func() {
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("existing-ver"))
				uploadReleaseOpts.Stemcell = boshdir.NewOSVersionSlug("ubuntu-trusty", "3421")

				director.HasReleaseReturns(true, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(0))

				name, version, stemcell := director.HasReleaseArgsForCall(0)
				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("existing-ver"))
				Expect(stemcell).To(Equal(boshdir.NewOSVersionSlug("ubuntu-trusty", "3421")))

				Expect(ui.Said).To(Equal(
					[]string{"Release 'existing-name/existing-ver' for stemcell 'ubuntu-trusty/3421' already exists."}))
			})

			It("uploads release if name and version does not match existing release", func() {
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("existing-ver"))

				director.HasReleaseReturns(false, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(1))

				url, sha1, rebase, fix := director.UploadReleaseURLArgsForCall(0)
				Expect(url).To(Equal("https://some-file.tzg"))
				Expect(sha1).To(Equal(""))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeFalse())

				name, version, stemcell := director.HasReleaseArgsForCall(0)
				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("existing-ver"))
				Expect(stemcell).To(Equal(boshdir.OSVersionSlug{}))

				Expect(ui.Said).To(BeEmpty())
			})

			It("returns error if checking for release existence fails", func() {
				director.HasReleaseReturns(false, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.UploadReleaseURLCallCount()).To(Equal(0))
			})

			It("returns error if uploading release failed", func() {
				director.UploadReleaseURLReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when url is a local file (file or no prefix)", func() {
			var (
				release *fakerel.FakeRelease
			)

			BeforeEach(func() {
				uploadReleaseOpts.Args.URL = "./some-file.tgz"

				release = &fakerel.FakeRelease{
					NameStub: func() string { return "rel" },
					ManifestStub: func() boshman.Manifest {
						return boshman.Manifest{Name: "rel"}
					},
				}
			})

			It("returns an error if reader is nil", func() {
				command = cmd.NewUploadReleaseCmd(nil, nil, director, nil, nil, nil, ui)

				err := command.Run(uploadReleaseOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot upload non-remote release"))
			})

			It("uploads given release", func() {
				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("./some-file.tgz"))
					return release, nil
				}

				director.MatchPackagesStub = func(manifest interface{}, compiled bool) ([]string, error) {
					Expect(manifest).To(Equal(boshman.Manifest{Name: "rel"}))
					Expect(compiled).To(BeFalse())
					return []string{"skip-pkg1-fp"}, nil
				}

				releaseWriter.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(Equal([]string{"skip-pkg1-fp"}))
					return "/archive-path", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.MatchPackagesCallCount()).To(Equal(1))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))

				file, rebase, fix := director.UploadReleaseFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("/archive-path"))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeFalse())
			})
			It("does not upload release if name and version match existing release", func() {
				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("./some-file.tgz"))
					return release, nil
				}
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("existing-ver"))
				director.HasReleaseReturns(true, nil)
				err := act()
				Expect(err).ToNot(HaveOccurred())

				name, version, stemcell := director.HasReleaseArgsForCall(0)
				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("existing-ver"))
				Expect(stemcell).To(Equal(boshdir.OSVersionSlug{}))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(0))
				Expect(ui.Said).To(Equal(
					[]string{"Release 'existing-name/existing-ver' already exists."}))
			})
			It("does upload a release if name and version match but exported from does not", func() {
				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("./some-file.tgz"))
					return release, nil
				}
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("existing-ver"))
				uploadReleaseOpts.Stemcell = boshdir.NewOSVersionSlug("ubuntu-xenial", "621.176")

				director.ReleaseHasCompiledPackageReturnsOnCall(1, false, nil)
				err := act()
				Expect(err).ToNot(HaveOccurred())
				name, version, stemcell := director.HasReleaseArgsForCall(0)
				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("existing-ver"))
				Expect(stemcell).To(Equal(boshdir.NewOSVersionSlug("ubuntu-xenial", "621.176")))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))
			})

			It("does upload a release if url points to a folder and version is create", func() {
				uploadReleaseOpts.Args.URL = "./some-folder"
				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("./some-folder"))
					return release, nil
				}
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("create"))
				uploadReleaseOpts.Stemcell = boshdir.NewOSVersionSlug("ubuntu-xenial", "621.176")
				err := act()
				Expect(err).ToNot(HaveOccurred())
				name, version, stemcell := director.HasReleaseArgsForCall(0)

				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("create"))
				Expect(stemcell).To(Equal(boshdir.NewOSVersionSlug("ubuntu-xenial", "621.176")))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))
			})

			It("clean up release", func() {
				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("./some-file.tgz"))
					return release, nil
				}

				releaseWriter.WriteStub = func(rel boshrel.Release, _ []string) (string, error) {
					Expect(rel).To(Equal(release))
					return "/archive-path", nil
				}

				removedFiles := []string{}

				fs.RemoveAllStub = func(path string) error {
					removedFiles = append(removedFiles, path)
					return nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(release.CleanUpCallCount()).To(Equal(1))
				Expect(removedFiles).To(Equal([]string{"/archive-path"}))
			})

			It("uploads given release with a fix flag hence does not filter out any packages", func() {
				uploadReleaseOpts.Fix = true

				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("./some-file.tgz"))
					return release, nil
				}

				releaseWriter.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(BeEmpty())
					return "/archive-path", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.MatchPackagesCallCount()).To(Equal(0))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))

				file, rebase, fix := director.UploadReleaseFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("/archive-path"))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeTrue())
			})

			It("returns error if opening file fails", func() {
				releaseReader.ReadReturns(release, nil)

				archive.FileStub = func() (boshdir.UploadFile, error) {
					return nil, errors.New("fake-err")
				}

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.UploadReleaseFileCallCount()).To(Equal(0))
			})

			It("returns error if uploading release failed", func() {
				releaseReader.ReadReturns(release, nil)
				director.UploadReleaseFileReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when url is a git repo", func() {
			var (
				release *fakerel.FakeRelease
			)

			BeforeEach(func() {
				// Command's --dir flag is not used
				uploadReleaseOpts.Args.URL = "git://./some-repo"
				uploadReleaseOpts.Directory = opts.DirOrCWDArg{Path: "/dir-that-does-not-matter"}

				// Destination for git clone
				fs.TempDirDir = "/dir"

				release = &fakerel.FakeRelease{
					NameStub: func() string { return "rel" },
					ManifestStub: func() boshman.Manifest {
						return boshman.Manifest{Name: "rel"}
					},
				}
			})

			It("returns an error if reader is nil", func() {
				command = cmd.NewUploadReleaseCmd(nil, nil, director, nil, cmdRunner, fs, ui)

				err := command.Run(uploadReleaseOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Cannot upload non-remote release"))
			})

			It("uploads given release", func() {
				uploadReleaseOpts.Name = "rel1"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("1.1"))
				afterClone := false

				cmdRunner.SetCmdCallback("git clone git://./some-repo --depth 1 /dir", func() {
					afterClone = true
				})

				releaseDir.FindReleaseStub = func(name string, version semver.Version) (boshrel.Release, error) {
					Expect(afterClone).To(BeTrue())
					Expect(name).To(Equal("rel1"))
					Expect(version).To(Equal(semver.MustNewVersionFromString("1.1")))
					return release, nil
				}

				director.MatchPackagesStub = func(manifest interface{}, compiled bool) ([]string, error) {
					Expect(manifest).To(Equal(boshman.Manifest{Name: "rel"}))
					Expect(compiled).To(BeFalse())
					return []string{"skip-pkg1-fp"}, nil
				}

				releaseWriter.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(Equal([]string{"skip-pkg1-fp"}))
					return "/archive-path", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.MatchPackagesCallCount()).To(Equal(1))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))

				file, rebase, fix := director.UploadReleaseFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("/archive-path"))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeFalse())
			})

			It("uploads given release with a fix flag hence does not filter out any packages", func() {
				uploadReleaseOpts.Fix = true

				releaseDir.FindReleaseStub = func(name string, version semver.Version) (boshrel.Release, error) {
					Expect(name).To(Equal(""))
					Expect(version).To(Equal(semver.Version{}))
					return release, nil
				}

				releaseWriter.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(BeEmpty())
					return "/archive-path", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.MatchPackagesCallCount()).To(Equal(0))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))

				file, rebase, fix := director.UploadReleaseFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("/archive-path"))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeTrue())
			})

			It("does not upload release if name and version match existing release", func() {
				uploadReleaseOpts.Name = "existing-name"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("existing-ver"))

				director.HasReleaseReturns(true, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.UploadReleaseURLCallCount()).To(Equal(0))

				name, version, stemcell := director.HasReleaseArgsForCall(0)
				Expect(name).To(Equal("existing-name"))
				Expect(version).To(Equal("existing-ver"))
				Expect(stemcell).To(Equal(boshdir.OSVersionSlug{}))

				Expect(ui.Said).To(Equal(
					[]string{"Release 'existing-name/existing-ver' already exists."}))
			})

			It("returns error if opening file fails", func() {
				releaseDir.FindReleaseReturns(release, nil)

				archive.FileStub = func() (boshdir.UploadFile, error) {
					return nil, errors.New("fake-err")
				}

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.UploadReleaseFileCallCount()).To(Equal(0))
			})

			It("returns error if creating temporary director failed", func() {
				fs.TempDirError = errors.New("fake-err")

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if git cloning failed", func() {
				cmdRunner.AddCmdResult("git clone git://./some-repo --depth 1 /dir", fakesys.FakeCmdResult{
					Error: errors.New("fake-err"),
				})

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if uploading release failed", func() {
				releaseDir.FindReleaseReturns(release, nil)
				director.UploadReleaseFileReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when url is empty", func() {
			var (
				release *fakerel.FakeRelease
			)

			BeforeEach(func() {
				uploadReleaseOpts.Args.URL = ""

				release = &fakerel.FakeRelease{
					NameStub: func() string { return "rel" },
					ManifestStub: func() boshman.Manifest {
						return boshman.Manifest{Name: "rel"}
					},
					IsCompiledStub: func() bool { return true },
				}
			})

			It("uploads found release based on name and version", func() {
				uploadReleaseOpts.Name = "rel1"
				uploadReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("1.1"))

				releaseDir.FindReleaseStub = func(name string, version semver.Version) (boshrel.Release, error) {
					Expect(name).To(Equal("rel1"))
					Expect(version).To(Equal(semver.MustNewVersionFromString("1.1")))
					return release, nil
				}

				director.MatchPackagesStub = func(manifest interface{}, compiled bool) ([]string, error) {
					Expect(manifest).To(Equal(boshman.Manifest{Name: "rel"}))
					Expect(compiled).To(BeTrue())
					return []string{"skip-pkg1-fp"}, nil
				}

				releaseWriter.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(Equal([]string{"skip-pkg1-fp"}))
					return "/archive-path", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.MatchPackagesCallCount()).To(Equal(1))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))

				file, rebase, fix := director.UploadReleaseFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("/archive-path"))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeFalse())
			})

			It("uploads given release with a fix flag and does not try to repack release", func() {
				uploadReleaseOpts.Fix = true

				releaseDir.FindReleaseReturns(release, nil)

				releaseWriter.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(BeEmpty())
					return "/archive-path", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.MatchPackagesCallCount()).To(Equal(0))
				Expect(director.UploadReleaseFileCallCount()).To(Equal(1))

				file, rebase, fix := director.UploadReleaseFileArgsForCall(0)
				Expect(file.(*fakesys.FakeFile).Name()).To(Equal("/archive-path"))
				Expect(rebase).To(BeFalse())
				Expect(fix).To(BeTrue())
			})

			It("returns error if finding release fails", func() {
				releaseDir.FindReleaseReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(director.UploadReleaseFileCallCount()).To(Equal(0))
			})

			It("returns error if uploading release failed", func() {
				releaseDir.FindReleaseReturns(release, nil)
				director.UploadReleaseFileReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
