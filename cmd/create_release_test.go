package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("CreateReleaseCmd", func() {
	var (
		releaseReader *fakerel.FakeReader
		releaseDir    *fakereldir.FakeReleaseDir
		ui            *fakeui.FakeUI
		fakeFS        *fakesys.FakeFileSystem
		fakeWriter    *fakerel.FakeWriter
		command       cmd.CreateReleaseCmd
	)

	BeforeEach(func() {
		releaseReader = &fakerel.FakeReader{}
		releaseDir = &fakereldir.FakeReleaseDir{}

		releaseDirFactory := func(dir opts.DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir) {
			Expect(dir).To(Equal(opts.DirOrCWDArg{Path: "/dir"}))
			return releaseReader, releaseDir
		}

		fakeWriter = &fakerel.FakeWriter{}
		fakeFS = fakesys.NewFakeFileSystem()
		ui = &fakeui.FakeUI{}
		command = cmd.NewCreateReleaseCmd(releaseDirFactory, fakeWriter, fakeFS, ui)
	})

	Describe("Run", func() {
		var (
			createReleaseOpts opts.CreateReleaseOpts
			release           *fakerel.FakeRelease
		)

		BeforeEach(func() {
			createReleaseOpts = opts.CreateReleaseOpts{
				Directory: opts.DirOrCWDArg{Path: "/dir"},
			}

			release = &fakerel.FakeRelease{
				NameStub:               func() string { return "rel" },
				VersionStub:            func() string { return "ver" },
				CommitHashWithMarkStub: func(string) string { return "commit" },

				SetNameStub:       func(name string) { release.NameReturns(name) },
				SetVersionStub:    func(ver string) { release.VersionReturns(ver) },
				NoCompressionStub: func() bool { return false },
			}
		})

		act := func() error {
			_, err := command.Run(createReleaseOpts)
			return err
		}

		Context("when manifest path is provided", func() {
			BeforeEach(func() {
				createReleaseOpts.Args.Manifest = opts.FileBytesWithPathArg{Path: "/manifest-path"}

				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("/manifest-path"))
					return release, nil
				}
			})

			It("builds release and release archive based on manifest path", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("rel"),
							boshtbl.NewValueString("ver"),
							boshtbl.NewValueString("commit"),
						},
					},
					Transpose: true,
				}))
			})

			It("returns error if reading manifest fails", func() {
				releaseReader.ReadReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			Context("with tarball", func() {
				BeforeEach(func() {
					createReleaseOpts.Tarball = opts.FileArg{ExpandedPath: "/tarball-destination.tgz"}
				})

				It("builds release and release archive based on manifest path", func() {
					fakeWriter.WriteStub = func(rel boshrel.Release, skipPkgs []string) (string, error) {
						Expect(rel).To(Equal(release))

						err := fakeFS.WriteFileString("/temp-tarball.tgz", "release content blah")
						Expect(err).ToNot(HaveOccurred())
						return "/temp-tarball.tgz", nil
					}

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
						Header: []boshtbl.Header{
							boshtbl.NewHeader("Name"),
							boshtbl.NewHeader("Version"),
							boshtbl.NewHeader("Commit Hash"),
							boshtbl.NewHeader("Archive"),
						},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("rel"),
								boshtbl.NewValueString("ver"),
								boshtbl.NewValueString("commit"),
								boshtbl.NewValueString("/tarball-destination.tgz"),
							},
						},
						Transpose: true,
					}))

					Expect(fakeFS.FileExists("/temp-tarball.tgz")).To(BeFalse())

					content, err := fakeFS.ReadFileString("/tarball-destination.tgz")
					Expect(err).ToNot(HaveOccurred())
					Expect(content).To(Equal("release content blah"))
				})

				It("interpolates release archive destination path with ((name)) and ((version))", func() {
					createReleaseOpts.Tarball = opts.FileArg{ExpandedPath: "/tarball-destination-((name))-((version)).tgz"}

					fakeWriter.WriteStub = func(rel boshrel.Release, skipPkgs []string) (string, error) {
						Expect(rel).To(Equal(release))

						err := fakeFS.WriteFileString("/temp-tarball.tgz", "release content blah")
						Expect(err).ToNot(HaveOccurred())
						return "/temp-tarball.tgz", nil
					}

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
						Header: []boshtbl.Header{
							boshtbl.NewHeader("Name"),
							boshtbl.NewHeader("Version"),
							boshtbl.NewHeader("Commit Hash"),
							boshtbl.NewHeader("Archive"),
						},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("rel"),
								boshtbl.NewValueString("ver"),
								boshtbl.NewValueString("commit"),
								boshtbl.NewValueString("/tarball-destination-rel-ver.tgz"),
							},
						},
						Transpose: true,
					}))

					Expect(fakeFS.FileExists("/temp-tarball.tgz")).To(BeFalse())

					content, err := fakeFS.ReadFileString("/tarball-destination-rel-ver.tgz")
					Expect(err).ToNot(HaveOccurred())
					Expect(content).To(Equal("release content blah"))
				})

				It("returns error if building release archive fails", func() {
					releaseReader.ReadReturns(release, nil)

					fakeWriter.WriteStub = func(rel boshrel.Release, skipPkgs []string) (string, error) {
						Expect(rel).To(Equal(release))
						return "", errors.New("fake-err")
					}

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("returns error moving the archive fails", func() {
					fakeWriter.WriteStub = func(rel boshrel.Release, skipPkgs []string) (string, error) {
						err := fakeFS.WriteFileString("/temp-tarball.tgz", "release content blah")
						Expect(err).ToNot(HaveOccurred())
						return "/temp-tarball.tgz", nil
					}

					fakeFS.RenameError = errors.New("fake-err")

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})
			})
		})

		Context("when manifest path is not provided", func() {
			It("builds release with default release name and next dev version", func() {
				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("next-dev+ver"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force, noCompression bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeFalse())
					return release, nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("default-rel-name"),
							boshtbl.NewValueString("next-dev+ver"),
							boshtbl.NewValueString("commit"),
						},
					},
					Transpose: true,
				}))
			})

			It("builds release with custom release name and version", func() {
				createReleaseOpts.Name = "custom-name"
				createReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("custom-ver"))

				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("1.1"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force, noCompression bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeFalse())
					return release, nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("custom-name"),
							boshtbl.NewValueString("custom-ver"),
							boshtbl.NewValueString("commit"),
						},
					},
					Transpose: true,
				}))
			})

			It("builds release forcefully with timestamp version", func() {
				createReleaseOpts.TimestampVersion = true
				createReleaseOpts.Force = true

				releaseDir.DefaultNameReturns("default-rel-name", nil)

				releaseDir.NextDevVersionStub = func(name string, timestamp bool) (semver.Version, error) {
					Expect(name).To(Equal("default-rel-name"))
					Expect(timestamp).To(BeTrue())
					return semver.MustNewVersionFromString("ts-ver"), nil
				}

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force, noCompression bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeTrue())
					return release, nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("default-rel-name"),
							boshtbl.NewValueString("ts-ver"),
							boshtbl.NewValueString("commit"),
						},
					},
					Transpose: true,
				}))
			})

			It("builds and then finalizes release", func() {
				createReleaseOpts.Final = true

				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("next-dev+ver"), nil)
				releaseDir.NextFinalVersionReturns(semver.MustNewVersionFromString("next-final+ver"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force, noCompression bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeFalse())
					return release, nil
				}

				releaseDir.FinalizeReleaseStub = func(rel boshrel.Release, force bool) error {
					Expect(rel).To(Equal(release))
					Expect(rel.Name()).To(Equal("default-rel-name"))
					Expect(rel.Version()).To(Equal("next-final+ver"))
					Expect(force).To(BeFalse())
					return nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("default-rel-name"),
							boshtbl.NewValueString("next-final+ver"),
							boshtbl.NewValueString("commit"),
						},
					},
					Transpose: true,
				}))
			})

			It("builds and then finalizes release with custom version", func() {
				createReleaseOpts.Final = true
				createReleaseOpts.Version = opts.VersionArg(semver.MustNewVersionFromString("custom-ver"))

				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("1.1"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force, noCompression bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeFalse())
					return release, nil
				}

				releaseDir.FinalizeReleaseStub = func(rel boshrel.Release, force bool) error {
					Expect(rel).To(Equal(release))
					Expect(rel.Name()).To(Equal("default-rel-name"))
					Expect(rel.Version()).To(Equal("custom-ver"))
					Expect(force).To(BeFalse())
					return nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("default-rel-name"),
							boshtbl.NewValueString("custom-ver"),
							boshtbl.NewValueString("commit"),
						},
					},
					Transpose: true,
				}))
			})

			It("builds release and archive if building archive is requested", func() {
				createReleaseOpts.Final = true
				createReleaseOpts.Tarball = opts.FileArg{ExpandedPath: "/archive-path"}

				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("next-dev+ver"), nil)
				releaseDir.NextFinalVersionReturns(semver.MustNewVersionFromString("next-final+ver"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force, noCompression bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeFalse())
					return release, nil
				}

				fakeWriter.WriteStub = func(rel boshrel.Release, skipPkgs []string) (string, error) {
					Expect(rel).To(Equal(release))

					err := fakeFS.WriteFileString("/temp-tarball.tgz", "release content blah")
					Expect(err).ToNot(HaveOccurred())
					return "/temp-tarball.tgz", nil
				}

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Commit Hash"),
						boshtbl.NewHeader("Archive"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("default-rel-name"),
							boshtbl.NewValueString("next-final+ver"),
							boshtbl.NewValueString("commit"),
							boshtbl.NewValueString("/archive-path"),
						},
					},
					Transpose: true,
				}))

				Expect(fakeFS.FileExists("/temp-tarball.tgz")).To(BeFalse())
				content, err := fakeFS.ReadFileString("/archive-path")
				Expect(err).ToNot(HaveOccurred())
				Expect(content).To(Equal("release content blah"))
			})

			It("returns error if retrieving default release name fails", func() {
				releaseDir.DefaultNameReturns("", errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Check that you're in the top-level of the release directory: fake-err"))
			})

			It("returns error if retrieving next dev version fails", func() {
				releaseDir.NextDevVersionReturns(semver.Version{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if retrieving next final version fails", func() {
				createReleaseOpts.Final = true

				releaseDir.BuildReleaseReturns(release, nil)
				releaseDir.NextFinalVersionReturns(semver.Version{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if building release archive fails", func() {
				createReleaseOpts.Tarball = opts.FileArg{ExpandedPath: "/tarball/dest/path.tgz"}

				fakeWriter.WriteStub = func(rel boshrel.Release, skipPkgs []string) (string, error) {
					return "", errors.New("fake-err")
				}

				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("next-dev+ver"), nil)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
