package releasedir_test

import (
	"errors"
	"os"
	"syscall"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshrel "github.com/cloudfoundry/bosh-init/release"
	fakerel "github.com/cloudfoundry/bosh-init/release/fakes"
	boshman "github.com/cloudfoundry/bosh-init/release/manifest"
	fakeres "github.com/cloudfoundry/bosh-init/release/resource/fakes"
	. "github.com/cloudfoundry/bosh-init/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-init/releasedir/fakes"
)

var _ = Describe("FSGenerator", func() {
	var (
		config        *fakereldir.FakeConfig
		gitRepo       *fakereldir.FakeGitRepo
		blobsDir      *fakereldir.FakeBlobsDir
		gen           *fakereldir.FakeGenerator
		devReleases   *fakereldir.FakeReleaseIndex
		finalReleases *fakereldir.FakeReleaseIndex
		finalIndicies boshrel.ArchiveIndicies
		reader        *fakerel.FakeReader
		writer        *fakerel.FakeWriter
		fs            *fakesys.FakeFileSystem
		releaseDir    FSReleaseDir
	)

	BeforeEach(func() {
		config = &fakereldir.FakeConfig{}
		gitRepo = &fakereldir.FakeGitRepo{}
		blobsDir = &fakereldir.FakeBlobsDir{}
		gen = &fakereldir.FakeGenerator{}
		devReleases = &fakereldir.FakeReleaseIndex{}
		finalReleases = &fakereldir.FakeReleaseIndex{}
		finalIndicies = boshrel.ArchiveIndicies{
			Jobs: &fakeres.FakeArchiveIndex{},
		}
		reader = &fakerel.FakeReader{}
		writer = &fakerel.FakeWriter{}
		fs = fakesys.NewFakeFileSystem()
		releaseDir = NewFSReleaseDir(
			"/dir",
			config,
			gitRepo,
			blobsDir,
			gen,
			devReleases,
			finalReleases,
			finalIndicies,
			reader,
			writer,
			fs,
		)
	})

	Describe("Init", func() {
		It("creates commont jobs, packages and src directories", func() {
			err := releaseDir.Init(true)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/dir/jobs")).To(BeTrue())
			Expect(fs.FileExists("/dir/packages")).To(BeTrue())
			Expect(fs.FileExists("/dir/src")).To(BeTrue())
		})

		It("returns error if creating common dirs fails", func() {
			fs.MkdirAllError = errors.New("fake-err")

			err := releaseDir.Init(true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("saves release name to directory base name", func() {
			err := releaseDir.Init(true)
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SaveFinalNameCallCount()).To(Equal(1))
			Expect(config.SaveFinalNameArgsForCall(0)).To(Equal("dir"))
		})

		It("returns error if saving final name fails", func() {
			config.SaveFinalNameReturns(errors.New("fake-err"))

			err := releaseDir.Init(true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("inits blobs", func() {
			err := releaseDir.Init(true)
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.InitCallCount()).To(Equal(1))
		})

		It("returns error if initing blobs fails", func() {
			blobsDir.InitReturns(errors.New("fake-err"))

			err := releaseDir.Init(true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("inits git repo if requested", func() {
			err := releaseDir.Init(true)
			Expect(err).ToNot(HaveOccurred())

			Expect(gitRepo.InitCallCount()).To(Equal(1))
		})

		It("does not init git repo if not requested", func() {
			err := releaseDir.Init(false)
			Expect(err).ToNot(HaveOccurred())

			Expect(gitRepo.InitCallCount()).To(Equal(0))
		})

		It("returns error if initing git repo fails", func() {
			gitRepo.InitReturns(errors.New("fake-err"))

			err := releaseDir.Init(true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("GenerateJob", func() {
		It("delegates to generator", func() {
			gen.GenerateJobStub = func(name string) error {
				Expect(name).To(Equal("job1"))
				return errors.New("fake-err")
			}
			Expect(releaseDir.GenerateJob("job1")).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("GeneratePackage", func() {
		It("delegates to generator", func() {
			gen.GeneratePackageStub = func(name string) error {
				Expect(name).To(Equal("job1"))
				return errors.New("fake-err")
			}
			Expect(releaseDir.GeneratePackage("job1")).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("ResetRelease", func() {
		It("removes .dev_builds and dev_releases", func() {
			fs.MkdirAll("/dev/.dev_builds/sub-dir", os.ModePerm)
			fs.MkdirAll("/dev/dev_releases/sub-dir", os.ModePerm)

			err := releaseDir.Reset()
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/dev/.dev_builds")).To(BeFalse())
			Expect(fs.FileExists("/dev/dev_releases")).To(BeFalse())
		})

		It("returns error when deleting directory fails", func() {
			fs.RemoveAllStub = func(_ string) error { return errors.New("fake-err") }

			err := releaseDir.Reset()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("DefaultName", func() {
		It("delegates to config", func() {
			config.FinalNameReturns("name", errors.New("fake-err"))

			name, err := releaseDir.DefaultName()
			Expect(name).To(Equal("name"))
			Expect(err).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("NextFinalVersion", func() {
		It("returns incremented last final version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.2").String()))
		})

		It("returns '0' if there are no versions so that when it's finalized it will be incremented to '1'", func() {
			finalReleases.LastVersionReturns(nil, nil)

			ver, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("0").String()))
		})

		It("returns error if cannot find out last version", func() {
			finalReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if incrementing fails", func() {
			lastVer := semver.MustNewVersionFromString("a")
			finalReleases.LastVersionReturns(&lastVer, nil)

			_, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incrementing last final version"))
		})
	})

	Describe("NextDevVersion", func() {
		It("returns incremented last final version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.1").String()))
		})

		It("returns incremented last dev version for specific release name", func() {
			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.2").String()))
		})

		It("returns incremented greater dev version compared to final version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.2").String()))
		})

		It("returns incremented greater final version compared to dev version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.2")
				return &lastVer, nil
			}

			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.2+dev.1").String()))
		})

		It("returns '0+dev.1' if there are no dev or final versions", func() {
			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("0+dev.1").String()))
		})

		It("returns error if cannot find out last dev version", func() {
			devReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot find out last final version", func() {
			finalReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if incrementing fails", func() {
			lastVer := semver.MustNewVersionFromString("1+a")
			finalReleases.LastVersionReturns(&lastVer, nil)

			_, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incrementing last dev version"))
		})
	})

	Describe("LastRelease", func() {
		var (
			expectedRelease *fakerel.FakeRelease
		)

		BeforeEach(func() {
			config.FinalNameReturns("rel1", nil)

			expectedRelease = &fakerel.FakeRelease{
				NameStub: func() string { return "rel1" },
			}
		})

		It("returns last final release for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			finalReleases.ManifestPathStub = func(name, ver string) string {
				Expect(name).To(Equal("rel1"))
				Expect(ver).To(Equal("1.1"))
				return "manifest-path"
			}

			reader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("manifest-path"))
				return expectedRelease, nil
			}

			release, err := releaseDir.LastRelease()
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(expectedRelease))
		})

		It("returns last dev release for specific release name", func() {
			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			devReleases.ManifestPathStub = func(name, ver string) string {
				Expect(name).To(Equal("rel1"))
				Expect(ver).To(Equal("1.1+dev.1"))
				return "manifest-path"
			}

			reader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("manifest-path"))
				return expectedRelease, nil
			}

			release, err := releaseDir.LastRelease()
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(expectedRelease))
		})

		It("returns greater dev release compared to final release for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			devReleases.ManifestPathStub = func(name, ver string) string {
				Expect(name).To(Equal("rel1"))
				Expect(ver).To(Equal("1.1+dev.1"))
				return "manifest-path"
			}

			reader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("manifest-path"))
				return expectedRelease, nil
			}

			release, err := releaseDir.LastRelease()
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(expectedRelease))
		})

		It("returns greater final release compared to dev release for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.2")
				return &lastVer, nil
			}

			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			finalReleases.ManifestPathStub = func(name, ver string) string {
				Expect(name).To(Equal("rel1"))
				Expect(ver).To(Equal("1.2"))
				return "manifest-path"
			}

			reader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("manifest-path"))
				return expectedRelease, nil
			}

			release, err := releaseDir.LastRelease()
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(expectedRelease))
		})

		It("returns error if there are no dev or final versions", func() {
			_, err := releaseDir.LastRelease()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find at least one dev or final version"))
		})

		It("returns error if cannot find out last dev version", func() {
			devReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.LastRelease()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot find out last final version", func() {
			finalReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.LastRelease()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("retuns error if cannot determine final name", func() {
			config.FinalNameReturns("", errors.New("fake-err"))

			_, err := releaseDir.LastRelease()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("BuildRelease", func() {
		var (
			ver             semver.Version
			expectedRelease *fakerel.FakeRelease
		)

		BeforeEach(func() {
			ver = semver.MustNewVersionFromString("1.1")

			expectedRelease = &fakerel.FakeRelease{
				NameStub: func() string { return "rel1" },
				ManifestStub: func() boshman.Manifest {
					return boshman.Manifest{Name: "rel1"}
				},
			}
		})

		It("builds release", func() {
			var ops []string

			gitRepo.MustNotBeDirtyStub = func(force bool) (bool, error) {
				ops = append(ops, "dirty")
				return true, nil
			}

			gitRepo.LastCommitSHAReturns("commit", nil)

			blobsDir.DownloadBlobsStub = func() error {
				ops = append(ops, "blobs")
				return nil
			}

			reader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("/dir"))
				ops = append(ops, "read")
				return expectedRelease, nil
			}

			devReleases.AddStub = func(manifest boshman.Manifest) error {
				Expect(manifest).To(Equal(boshman.Manifest{Name: "rel1"}))
				ops = append(ops, "manifest")
				return nil
			}

			release, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(expectedRelease))

			Expect(expectedRelease.SetNameArgsForCall(0)).To(Equal("rel1"))
			Expect(expectedRelease.SetVersionArgsForCall(0)).To(Equal("1.1"))
			Expect(expectedRelease.SetCommitHashArgsForCall(0)).To(Equal("commit"))
			Expect(expectedRelease.SetUncommittedChangesArgsForCall(0)).To(BeTrue())

			Expect(ops).To(Equal([]string{"dirty", "blobs", "read", "manifest"}))
		})

		It("returns error if git is dirty and force is not set", func() {
			gitRepo.MustNotBeDirtyReturns(true, errors.New("dirty"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(Equal(errors.New("dirty")))
		})

		It("returns error if last commit cannot be retrieved", func() {
			gitRepo.LastCommitSHAReturns("", errors.New("fake-err"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})

		It("returns error if reading release", func() {
			reader.ReadReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if adding dev release fails", func() {
			reader.ReadReturns(expectedRelease, nil)
			devReleases.AddReturns(errors.New("fake-err"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("FinalizeRelease", func() {
		var (
			release *fakerel.FakeRelease
		)

		BeforeEach(func() {
			release = &fakerel.FakeRelease{
				NameStub:    func() string { return "rel1" },
				VersionStub: func() string { return "ver1" },
				ManifestStub: func() boshman.Manifest {
					return boshman.Manifest{Name: "rel1"}
				},
			}
		})

		It("finalizes release", func() {
			var ops []string

			gitRepo.MustNotBeDirtyStub = func(force bool) (bool, error) {
				ops = append(ops, "dirty")
				return true, nil
			}

			finalReleases.ContainsStub = func(rel boshrel.Release) (bool, error) {
				Expect(rel).To(Equal(release))
				ops = append(ops, "check")
				return false, nil
			}

			release.FinalizeStub = func(indicies boshrel.ArchiveIndicies) error {
				Expect(indicies.Jobs).To(Equal(finalIndicies.Jobs)) // unique check
				ops = append(ops, "finalize")
				return nil
			}

			finalReleases.AddStub = func(manifest boshman.Manifest) error {
				Expect(manifest).To(Equal(boshman.Manifest{Name: "rel1"}))
				ops = append(ops, "manifest")
				return nil
			}

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(ops).To(Equal([]string{"dirty", "check", "finalize", "manifest"}))
		})

		It("returns error if git is dirty and force is not set", func() {
			gitRepo.MustNotBeDirtyReturns(true, errors.New("dirty"))

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(Equal(errors.New("dirty")))
		})

		It("returns error if checking for a final release fails", func() {
			finalReleases.ContainsReturns(false, errors.New("fake-err"))

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})

		It("returns error if final release index already contains this name/ver", func() {
			finalReleases.ContainsReturns(true, nil)

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Release 'rel1' version 'ver1' already exists"))
		})

		It("returns error if adding final release fails", func() {
			finalReleases.AddReturns(errors.New("fake-err"))

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("BuildReleaseArchive", func() {
		var (
			release *fakerel.FakeRelease
		)

		BeforeEach(func() {
			release = &fakerel.FakeRelease{
				NameStub:    func() string { return "rel1" },
				VersionStub: func() string { return "ver1" },
			}
		})

		It("adds release archive to final releases", func() {
			writer.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
				Expect(rel).To(Equal(release))
				Expect(pkgFpsToSkip).To(BeNil())
				fs.WriteFileString("/tmp/archive-path", "archive")
				return "/tmp/archive-path", nil
			}

			finalReleases.ArchivePathStub = func(rel boshrel.Release) (string, error) {
				Expect(rel).To(Equal(release))
				return "/tmp/final-archive-path", nil
			}

			path, err := releaseDir.BuildReleaseArchive(release)
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("/tmp/final-archive-path"))

			Expect(fs.FileExists("/tmp/archive-path")).To(BeFalse())
			Expect(fs.ReadFileString("/tmp/final-archive-path")).To(Equal("archive"))
		})

		It("adds release archive to dev releases", func() {
			release.VersionReturns("ver1+dev.1") // dev makes it a dev release

			writer.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
				Expect(rel).To(Equal(release))
				Expect(pkgFpsToSkip).To(BeNil())
				fs.WriteFileString("/tmp/archive-path", "archive")
				return "/tmp/archive-path", nil
			}

			devReleases.ArchivePathStub = func(rel boshrel.Release) (string, error) {
				Expect(rel).To(Equal(release))
				return "/tmp/final-archive-path", nil
			}

			path, err := releaseDir.BuildReleaseArchive(release)
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(Equal("/tmp/final-archive-path"))

			Expect(fs.FileExists("/tmp/archive-path")).To(BeFalse())
			Expect(fs.ReadFileString("/tmp/final-archive-path")).To(Equal("archive"))
		})

		Context("when archive writing fails", func() {
			It("returns error", func() {
				writer.WriteReturns("", errors.New("fake-err"))

				_, err := releaseDir.BuildReleaseArchive(release)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when obtaining final release archive path fails", func() {
			It("returns error", func() {
				finalReleases.ArchivePathReturns("", errors.New("fake-err"))

				_, err := releaseDir.BuildReleaseArchive(release)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when moving archive across devices", func() {
			BeforeEach(func() {
				fs.RenameError = syscall.Errno(0x12)
			})

			It("moves release archive successfully", func() {
				writer.WriteStub = func(rel boshrel.Release, pkgFpsToSkip []string) (string, error) {
					Expect(rel).To(Equal(release))
					Expect(pkgFpsToSkip).To(BeNil())
					fs.WriteFileString("/tmp/archive-path", "archive")
					return "/tmp/archive-path", nil
				}

				finalReleases.ArchivePathStub = func(rel boshrel.Release) (string, error) {
					Expect(rel).To(Equal(release))
					return "/tmp/final-archive-path", nil
				}

				path, err := releaseDir.BuildReleaseArchive(release)
				Expect(err).ToNot(HaveOccurred())
				Expect(path).To(Equal("/tmp/final-archive-path"))

				Expect(fs.FileExists("/tmp/archive-path")).To(BeFalse())
				Expect(fs.ReadFileString("/tmp/final-archive-path")).To(Equal("archive"))
			})

			Context("when copying across devices fails", func() {
				It("returns error if moving archive to final destination fails", func() {
					fs.CopyFileError = errors.New("copy-err")

					_, err := releaseDir.BuildReleaseArchive(release)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("copy-err"))
				})
			})
		})

		Context("moving archive to final destination fails for unknown reason", func() {
			It("returns error", func() {
				fs.RenameError = errors.New("fake-err")

				_, err := releaseDir.BuildReleaseArchive(release)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
