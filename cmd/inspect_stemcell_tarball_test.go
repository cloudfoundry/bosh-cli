package cmd_test

import (
	"errors"
	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InspectStemcellTarballCmd", func() {
	Describe("Run", func() {
		var (
			fs               *fakesys.FakeFileSystem
			archive          *fakedir.FakeStemcellArchive
			command          InspectStemcellTarballCmd
			ui               *fakeui.FakeUI
			opts             InspectStemcellTarballOpts
			stemcellMetadata boshdir.StemcellMetadata
		)

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			archive = &fakedir.FakeStemcellArchive{}
			stemcellMetadata = boshdir.StemcellMetadata{Name: "example-name", OS: "example-os", Version: "example.version"}

			stemcellArchiveFactory := func(path string) boshdir.StemcellArchive {
				if archive.FileStub == nil {
					archive.FileStub = func() (boshdir.UploadFile, error) {
						return fakesys.NewFakeFile(path, fs), nil
					}
				}
				return archive
			}

			opts = InspectStemcellTarballOpts{}
			ui = &fakeui.FakeUI{}

			command = NewInspectStemcellTarballCmd(stemcellArchiveFactory, ui)
		})

		It("returns a table with name, os, and version", func() {
			archive.InfoReturns(stemcellMetadata, nil)

			err := command.Run(opts)
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "stemcell-metadata",

				Header: []boshtbl.Header{
					boshtbl.NewHeader("Name"),
					boshtbl.NewHeader("OS"),
					boshtbl.NewHeader("Version"),
				},

				SortBy: []boshtbl.ColumnSort{
					{Column: 0, Asc: true},
				},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("example-name"),
						boshtbl.NewValueString("example-os"),
						boshtbl.NewValueString("example.version"),
					},
				},
			}))
		})

		It("returns error if retrieving stemcell archive info fails", func() {
			archive.InfoReturns(boshdir.StemcellMetadata{}, errors.New("fake-err"))

			err := command.Run(opts)
			Expect(err).To(HaveOccurred())
		})
	})
})
