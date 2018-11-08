package cmd_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
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
			stemcellMetadata = boshdir.StemcellMetadata{
				Name:    "example-name",
				OS:      "example-os",
				Version: "example.version",
				CloudProperties: biproperty.Map{
					"infrastructure": "example-infrastructure",
				},
			}

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

		Context("when infrastructure is known", func() {
			It("returns a table with name, os, version, and infrastructure", func() {
				archive.InfoReturns(stemcellMetadata, nil)

				err := command.Run(opts)
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "stemcell-metadata",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("OS"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Infrastructure"),
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("example-name"),
							boshtbl.NewValueString("example-os"),
							boshtbl.NewValueString("example.version"),
							boshtbl.NewValueString("example-infrastructure"),
						},
					},
				}))
			})
		})

		Context("when infrastructure is unknown", func() {

			BeforeEach(func() {
				stemcellMetadata.CloudProperties["infrastructure"] = nil
			})

			It("returns a table with name, os, version, and infrastructure", func() {
				archive.InfoReturns(stemcellMetadata, nil)

				err := command.Run(opts)
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "stemcell-metadata",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("OS"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("Infrastructure"),
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("example-name"),
							boshtbl.NewValueString("example-os"),
							boshtbl.NewValueString("example.version"),
							boshtbl.NewValueString("unknown"),
						},
					},
				}))
			})
		})

		It("returns error if retrieving stemcell archive info fails", func() {
			archive.InfoReturns(boshdir.StemcellMetadata{}, errors.New("fake-err"))

			err := command.Run(opts)
			Expect(err).To(HaveOccurred())
		})
	})
})
