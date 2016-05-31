package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	fakerel "github.com/cloudfoundry/bosh-init/release/fakes"
	boshjob "github.com/cloudfoundry/bosh-init/release/job"
	boshpkg "github.com/cloudfoundry/bosh-init/release/pkg"
	. "github.com/cloudfoundry/bosh-init/release/resource"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("ReleaseTables", func() {
	var (
		ui *fakeui.FakeUI
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
	})

	Describe("Print", func() {
		var (
			release *fakerel.FakeRelease
		)

		BeforeEach(func() {
			job := boshjob.NewJob(NewResourceWithBuiltArchive(
				"job-name", "job-fp", "job-path", "job-sha1"))

			pkg := boshpkg.NewPackage(NewResourceWithBuiltArchive(
				"pkg-name", "pkg-fp", "pkg-path", "pkg-sha1"), nil)

			release = &fakerel.FakeRelease{
				NameStub:    func() string { return "rel" },
				VersionStub: func() string { return "ver" },

				CommitHashWithMarkStub: func(string) string { return "commit" },

				JobsStub:     func() []*boshjob.Job { return []*boshjob.Job{job} },
				PackagesStub: func() []*boshpkg.Package { return []*boshpkg.Package{pkg} },
			}
		})

		It("shows info about release with archive path", func() {
			ReleaseTables{Release: release, ArchivePath: "/archive-path"}.Print(ui)

			Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
				Rows: [][]boshtbl.Value{
					{
						boshtbl.ValueString{"Name"},
						boshtbl.ValueString{"rel"},
					},
					{
						boshtbl.ValueString{"Version"},
						boshtbl.ValueString{"ver"},
					},
					{
						boshtbl.ValueString{"Commit Hash"},
						boshtbl.ValueString{"commit"},
					},
					{
						boshtbl.ValueString{"Archive"},
						boshtbl.ValueString{"/archive-path"},
					},
				},
			}))

			Expect(ui.Tables[1]).To(Equal(boshtbl.Table{
				Content: "jobs",
				Header:  []string{"Job", "SHA1", "Packages"},
				SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
				Rows: [][]boshtbl.Value{
					{
						boshtbl.ValueString{"job-name/job-fp"},
						boshtbl.ValueString{"job-sha1"},
						boshtbl.ValueString{""},
					},
				},
			}))

			Expect(ui.Tables[2]).To(Equal(boshtbl.Table{
				Content: "packages",
				Header:  []string{"Package", "SHA1", "Dependencies"},
				SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
				Rows: [][]boshtbl.Value{
					{
						boshtbl.ValueString{"pkg-name/pkg-fp"},
						boshtbl.ValueString{"pkg-sha1"},
						boshtbl.ValueString{""},
					},
				},
			}))
		})

		It("shows info about release without archive path", func() {
			ReleaseTables{Release: release}.Print(ui)

			Expect(ui.Tables[0]).To(Equal(boshtbl.Table{
				Rows: [][]boshtbl.Value{
					{
						boshtbl.ValueString{"Name"},
						boshtbl.ValueString{"rel"},
					},
					{
						boshtbl.ValueString{"Version"},
						boshtbl.ValueString{"ver"},
					},
					{
						boshtbl.ValueString{"Commit Hash"},
						boshtbl.ValueString{"commit"},
					},
				},
			}))
		})
	})
})
