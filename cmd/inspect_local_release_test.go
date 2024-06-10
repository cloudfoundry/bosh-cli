package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshjob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("InspectLocalReleaseCmd", func() {
	Describe("Run", func() {
		var (
			fakeRelease             *fakerel.FakeRelease
			releaseReader           *fakerel.FakeReader
			ui                      *fakeui.FakeUI
			inspectLocalReleaseOpts opts.InspectLocalReleaseOpts
			command                 cmd.InspectLocalReleaseCmd
		)

		BeforeEach(func() {
			fakeRelease = &fakerel.FakeRelease{}
			fakeRelease.NameReturns("rel")
			fakeRelease.VersionReturns("ver")
			fakeRelease.CommitHashWithMarkReturns("commit")

			job := boshjob.NewJob(NewResourceWithBuiltArchive(
				"job-name",
				"job-fp",
				"/job-resource-path",
				"job-digest",
			))
			job.PackageNames = []string{"pkg-1-name", "pkg-2-name"}

			pkg1 := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg-1-name",
				"pkg-1-fp",
				"/pkg-1-resource-path",
				"pkg-1-digest",
			), nil)
			pkg2 := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg-2-name",
				"pkg-2-fp",
				"/pkg-2-resource-path",
				"pkg-2-digest"),
				[]string{"pkg-1-name"},
			)
			err := pkg2.AttachDependencies([]*boshpkg.Package{pkg1})
			Expect(err).ToNot(HaveOccurred())

			err = job.AttachPackages([]*boshpkg.Package{pkg1, pkg2})
			Expect(err).ToNot(HaveOccurred())

			compiledPkg := boshpkg.NewCompiledPackageWithoutArchive(
				"compiled-pkg-name",
				"compiled-pkg-fp",
				"my-fancy-linux/1.33.7",
				"compiled-pkg-digest",
				[]string{"some-package"},
			)

			fakeRelease.JobsReturns([]*boshjob.Job{job})
			fakeRelease.PackagesReturns([]*boshpkg.Package{pkg1, pkg2})
			fakeRelease.CompiledPackagesReturns([]*boshpkg.CompiledPackage{compiledPkg})

			releaseReader = &fakerel.FakeReader{}
			releaseReader.ReadReturns(fakeRelease, nil)

			inspectLocalReleaseOpts = opts.InspectLocalReleaseOpts{
				Args: opts.InspectLocalReleaseArgs{
					PathToRelease: "/some/release.tgz",
				},
			}

			ui = &fakeui.FakeUI{}

			command = cmd.NewInspectLocalReleaseCmd(releaseReader, ui)
		})

		It("prints tables with release, job and package information", func() {
			err := command.Run(inspectLocalReleaseOpts)
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
						boshtbl.NewValueString("/some/release.tgz"),
					},
				},
				Transpose: true,
			}))

			Expect(ui.Tables[1]).To(Equal(boshtbl.Table{
				Content: "jobs",
				Header: []boshtbl.Header{
					boshtbl.NewHeader("Job"),
					boshtbl.NewHeader("Digest"),
					boshtbl.NewHeader("Packages"),
				},
				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("job-name/job-fp"),
						boshtbl.NewValueString("job-digest"),
						boshtbl.NewValueStrings([]string{"pkg-1-name", "pkg-2-name"}),
					},
				},
			}))

			var emptyNames []string

			Expect(ui.Tables[2]).To(Equal(boshtbl.Table{
				Content: "packages",
				Header: []boshtbl.Header{
					boshtbl.NewHeader("Package"),
					boshtbl.NewHeader("Digest"),
					boshtbl.NewHeader("Dependencies"),
					boshtbl.NewHeader("OS"),
					boshtbl.NewHeader("OS Version"),
				},
				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("pkg-1-name/pkg-1-fp"),
						boshtbl.NewValueString("pkg-1-digest"),
						boshtbl.NewValueStrings(emptyNames),
						boshtbl.NewValueString(""),
						boshtbl.NewValueString(""),
					},
					{
						boshtbl.NewValueString("pkg-2-name/pkg-2-fp"),
						boshtbl.NewValueString("pkg-2-digest"),
						boshtbl.NewValueStrings([]string{"pkg-1-name"}),
						boshtbl.NewValueString(""),
						boshtbl.NewValueString(""),
					},
					{
						boshtbl.NewValueString("compiled-pkg-name/compiled-pkg-fp"),
						boshtbl.NewValueString("compiled-pkg-digest"),
						boshtbl.NewValueStrings(nil),
						boshtbl.NewValueString("my-fancy-linux"),
						boshtbl.NewValueString("1.33.7"),
					},
				},
			}))
		})

		It("returns error if reading the release manifest fails", func() {
			releaseReader.ReadReturns(nil, errors.New("fake-err"))

			err := command.Run(inspectLocalReleaseOpts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
