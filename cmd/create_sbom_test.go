package cmd_test

import (
	"errors"

	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("CreateSbomCmd", func() {
	var (
		releaseReader *fakerel.FakeReader
		releaseDir    *fakereldir.FakeReleaseDir
		ui            *fakeui.FakeUI
		fakeFS        *fakesys.FakeFileSystem
		fakeWriter    *fakerel.FakeWriter
		compressor    *FakeCompressor
		command       CreateSbomCmd
	)

	BeforeEach(func() {
		releaseReader = &fakerel.FakeReader{}
		releaseDir = &fakereldir.FakeReleaseDir{}

		releaseDirFactory := func(dir DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir) {
			Expect(dir).To(Equal(DirOrCWDArg{Path: "/dir"}))
			return releaseReader, releaseDir
		}

		fakeWriter = &fakerel.FakeWriter{}
		fakeFS = fakesys.NewFakeFileSystem()
		compressor = NewFakeCompressor(fakeFS)
		ui = &fakeui.FakeUI{}
		command = NewCreateSbomCmd(releaseDirFactory, fakeWriter, compressor, fakeFS, ui)
	})

	Describe("Run", func() {
		var (
			opts    CreateSbomOpts
			release *fakerel.FakeRelease
		)

		BeforeEach(func() {
			opts = CreateSbomOpts{
				Directory: DirOrCWDArg{Path: "/dir"},
			}

			release = &fakerel.FakeRelease{
				NameStub:               func() string { return "rel" },
				VersionStub:            func() string { return "ver" },
				CommitHashWithMarkStub: func(string) string { return "commit" },

				SetNameStub:    func(name string) { release.NameReturns(name) },
				SetVersionStub: func(ver string) { release.VersionReturns(ver) },
			}
		})

		act := func() error {
			_, err := command.Run(opts)
			return err
		}

		Context("when manifest path is provided", func() {
			BeforeEach(func() {
				opts.Args.Manifest = FileBytesWithPathArg{Path: "/manifest-path"}

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
		})

		Context("when manifest path is not provided", func() {
			It("builds release with default release name and next dev version", func() {
				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("next-dev+ver"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force bool) (boshrel.Release, error) {
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
				opts.Name = "custom-name"
				opts.Version = VersionArg(semver.MustNewVersionFromString("custom-ver"))

				releaseDir.DefaultNameReturns("default-rel-name", nil)
				releaseDir.NextDevVersionReturns(semver.MustNewVersionFromString("1.1"), nil)

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force bool) (boshrel.Release, error) {
					release.SetName(name)
					release.SetVersion(version.String())
					Expect(force).To(BeFalse())
					return release, nil
				}

				err := act()
				Expect(err).To(HaveOccurred())

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
				opts.TimestampVersion = true
				opts.Force = true

				releaseDir.DefaultNameReturns("default-rel-name", nil)

				releaseDir.NextDevVersionStub = func(name string, timestamp bool) (semver.Version, error) {
					Expect(name).To(Equal("default-rel-name"))
					Expect(timestamp).To(BeTrue())
					return semver.MustNewVersionFromString("ts-ver"), nil
				}

				releaseDir.BuildReleaseStub = func(name string, version semver.Version, force bool) (boshrel.Release, error) {
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

			It("returns error if retrieving default release name fails", func() {
				releaseDir.DefaultNameReturns("", errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if retrieving next dev version fails", func() {
				releaseDir.NextDevVersionReturns(semver.Version{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

		})
	})
})
